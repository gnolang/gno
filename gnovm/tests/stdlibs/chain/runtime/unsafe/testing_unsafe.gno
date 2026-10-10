package unsafe

// native binding override: routes through the test stdlib's getRealm
// so testing.SetRealm / RealmFrames overrides take effect for
// unsafe.PreviousRealm() and unsafe.CurrentRealm() in test code.
func getRealm(height int) (addr string, pkgPath string)
