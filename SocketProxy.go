// +build linux

package transproxy

import (
	"fmt"
	"io"
	"net"
	"syscall"
	"time"

	"github.com/alexcesaro/log"
	"github.com/xiqingping/transproxy/proxy"
)

type sockaddr struct {
	family uint16
	data   [14]byte
}

const SO_ORIGINAL_DST = 80

func getOriginalDst(tcpConn *net.TCPConn) (addr net.TCPAddr, newTCPConn *net.TCPConn, err error) {
	newTCPConn = tcpConn
	// net.TCPConn.File() will cause the receiver's (clientConn) socket to be placed in blocking mode.
	// The workaround is to take the File returned by .File(), do getsockopt() to get the original
	// destination, then create a new *net.TCPConn by calling net.Conn.FileConn().  The new TCPConn
	// will be in non-blocking mode.  What a pain.
	connFile, err := tcpConn.File()
	if err != nil {
		return
	} else {
		tcpConn.Close()
	}

	// Get original destination
	// this is the only syscall in the Golang libs that I can find that returns 16 bytes
	// Example result: &{Multiaddr:[2 0 31 144 206 190 36 45 0 0 0 0 0 0 0 0] Interface:0}
	// port starts at the 3rd byte and is 2 bytes long (31 144 = port 8080)
	// IPv4 address starts at the 5th byte, 4 bytes long (206 190 36 45)
	mreq, err := syscall.GetsockoptIPv6Mreq(int(connFile.Fd()), syscall.IPPROTO_IP, SO_ORIGINAL_DST)
	if err != nil {
		return
	}
	newConn, err := net.FileConn(connFile)
	if err != nil {
		return
	}

	newTCPConn = newConn.(*net.TCPConn)
	connFile.Close()

	addr.IP = mreq.Multiaddr[4:8]
	addr.Port = int(mreq.Multiaddr[2])<<8 + int(mreq.Multiaddr[3])

	return
}

type SocketProxy struct {
	conn      *net.TCPConn
	dest      net.TCPAddr
	logger    log.Logger
	bl        *BlackList
	proxyConn net.Conn
	proxyDial proxy.Dialer
}

func NewSocketProxy(conn *net.TCPConn, bl *BlackList, proxyDial proxy.Dialer, logger log.Logger) (sp *SocketProxy, err error) {
	dest, conn, err := getOriginalDst(conn)
	if err != nil {
		conn.Close()
		return
	}

	sp = &SocketProxy{
		conn:      conn,
		dest:      dest,
		logger:    logger,
		bl:        bl,
		proxyDial: proxyDial,
	}
	return
}

func (sp *SocketProxy) String() string {
	return fmt.Sprintf("[%v->%v->%v]", sp.conn.RemoteAddr(), sp.conn.LocalAddr(), sp.dest.String())
}

func copyAndCloseReader(rc io.ReadCloser, w io.Writer) {
	io.Copy(w, rc)
	rc.Close()
}

func (sp *SocketProxy) Run() {
	var err error
	needProxy := false

	if sp.bl.Contains(sp.dest.IP) {
		sp.logger.Debug("Dest in black list")
		needProxy = true
	} else {
		sp.proxyConn, err = net.DialTimeout("tcp4", sp.dest.String(), time.Second*10)
		if err != nil {
			sp.logger.Warning(sp, "Direct dial to", sp.dest, "error:", err, ", try proxy")
			needProxy = true
		}
	}

	if needProxy {
		sp.proxyConn, err = sp.proxyDial.Dial("tcp", sp.dest.String())
		if err != nil {
			sp.logger.Warning(sp, "Proxy dial to", sp.dest, "error:", err)
			sp.conn.Close()
			sp.logger.Info(sp, "Finished")
			return
		}
		sp.logger.Debug("Add", sp.dest.IP, "to black list")
		sp.bl.Add(sp.dest.IP)
	}

	sp.logger.Info(sp, "Start proxy ...")
	go copyAndCloseReader(sp.proxyConn, sp.conn)
	copyAndCloseReader(sp.conn, sp.proxyConn)
	sp.logger.Info(sp, "Finished")
}
