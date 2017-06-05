// +build linux

package main

//#include <errno.h>
//#include <sys/types.h>
//#include <unistd.h>
//int GetErrno(void) {
//  return errno;
//}
import "C"

import (
	"fmt"
	"os/user"
	"strconv"
	"syscall"
)

func dropPrivileges(uidstr, gidstr string) error {
	if gidstr != "" {
		grp, err := user.LookupGroup(gidstr)
		if err != nil {
			return fmt.Errorf("LookupGroup(%v) %v", gidstr, syscall.Errno(C.GetErrno()))
		}
		id, _ := strconv.Atoi(grp.Gid)
		if 0 != C.setgid(C.__gid_t(id)) {
			return fmt.Errorf("Setgid(%v) %v", id, syscall.Errno(C.GetErrno()))
		}
	}

	if uidstr != "" {
		user, err := user.Lookup(uidstr)
		if err != nil {
			return fmt.Errorf("Lookup(%v) %v", uidstr, syscall.Errno(C.GetErrno()))
		}

		id, _ := strconv.Atoi(user.Uid)
		if 0 != C.setuid(C.__uid_t(id)) {
			return fmt.Errorf("Setuid(%v) %v", id, syscall.Errno(C.GetErrno()))
		}
	}

	return nil
}
