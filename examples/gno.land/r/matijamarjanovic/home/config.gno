package home

import (
        "errors"
        "std"
)

var (
        mainAddr   std.Address // tony's main address
        backupAddr std.Address // backup address

        errorInvalidAddr  string
        errorUnauthorized string
)

func init() {
        mainAddr = "g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"
        errorInvalidAddr = "config: invalid address"
        errorUnauthorized = "config: unauthorized"
}

func Address() std.Address {
        return mainAddr
}

func Backup() std.Address {
        return backupAddr
}

func SetAddress(newAddress std.Address) error {
        if !newAddress.IsValid() {
                return errors.New(errorInvalidAddr)
        }

        if err := checkAuthorized(); err != nil {
                return err
        }

        mainAddr = newAddress
        return nil
}

func SetBackup(newAddress std.Address) error {
        if !newAddress.IsValid() {
                return errors.New(errorInvalidAddr)
        }

        if err := checkAuthorized(); err != nil {
                return err
        }

        backupAddr = newAddress
        return nil
}

func checkAuthorized() error {
        caller := std.PrevRealm().Addr()
        if caller != mainAddr && caller != backupAddr {
                return errors.New(errorUnauthorized)
        }

        return nil
}

func AssertAuthorized() {
        caller := std.PrevRealm().Addr()
        if caller != mainAddr && caller != backupAddr {
                panic(errorUnauthorized)
        }
}