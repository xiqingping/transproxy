# transproxy
transparent proxy to cccccros the GFW, cooperate with iptables; also see redsocks: https://github.com/darkk/redsocks

thanks the proxy library from here: https://github.com/phuslu/goproxy/tree/master/httpproxy/proxy

1. compile
    ```bash
    go get -u github.com/xiqingping/transproxy/...
    ```

1. config the iptables

    ```bash
    # create TRANSPROXYD chain
    iptables -t nat -N TRANSPROXYD

    # packets from nobody(transproxyd run as nobody) do not redirect
    iptables -t nat -A TRANSPROXYD -p tcp -m owner --uid-owner nobody -j RETURN

    # packets to private net do not redirect
    iptables -t nat -A TRANSPROXYD -d 0.0.0.0/8 -j RETURN
    iptables -t nat -A TRANSPROXYD -d 10.0.0.0/8 -j RETURN
    iptables -t nat -A TRANSPROXYD -d 127.0.0.0/8 -j RETURN
    iptables -t nat -A TRANSPROXYD -d 169.254.0.0/16 -j RETURN
    iptables -t nat -A TRANSPROXYD -d 172.16.0.0/12 -j RETURN
    iptables -t nat -A TRANSPROXYD -d 192.168.0.0/16 -j RETURN
    iptables -t nat -A TRANSPROXYD -d 224.0.0.0/4 -j RETURN
    iptables -t nat -A TRANSPROXYD -d 240.0.0.0/4 -j RETURN

    # redirect all tcp packets to port 12345(transproxyd listen on this port)
    iptables -t nat -A TRANSPROXYD -p tcp -j REDIRECT --to-ports 12345

    # all output tcp packets jump to TRANSPROXYD chain
    iptables -t nat -A OUTPUT -p tcp  -j TRANSPROXYD

    ```

1. Run the transproxyd, pls ref the config file under github.com/xiqingping/transproxy/transproxyd
    ```
    sudo transproxyd -config path/to/transproxyd.toml
    ```

   - Uid or Gid in the config must satify to iptables config `iptables -t nat -A TRANSPROXYD -p tcp -m owner --uid-owner nobody -j RETURN`
   - ListenAddr in the config must satify to iptables config `iptables -t nat -A TRANSPROXYD -p tcp -j REDIRECT --to-ports 12345`
