package auth

import (
	"golang.org/x/sys/unix"
	"log"
	"net"
	"os/user"
)

func Allowed(conn *net.UnixConn, groups map[string]struct{}) bool {
	if conn == nil || groups == nil || len(groups) == 0 {
		return false
	}
	file, _ := conn.File()
	defer file.Close()

	var (
		err   error
		ucred *unix.Ucred
	)

	for {
		ucred, err = unix.GetsockoptUcred(int(file.Fd()), unix.SOL_SOCKET, unix.SO_PEERCRED)
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			log.Println(err)
			return false
		}
		break
	}
	u, err := user.LookupId(string(ucred.Uid))
	if err != nil {
		log.Println(err)
		return false
	}
	gids, err := u.GroupIds()
	if err != nil {
		log.Println(err)
		return false
	}
	for _, gid := range gids {
		if _, ok := groups[gid]; ok {
			return true
		}
	}
	return false
}
