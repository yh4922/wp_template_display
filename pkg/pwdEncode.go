package pkg

import (
	"github.com/dromara/dongle"
)

// 密码加密
func PwdEncode(pwd string) string {
	sign := dongle.Sign.FromString(pwd).ByBcrypt()
	return sign.ToRawString()
}

func PwdCompare(pwd string, hexPwd string) bool {
	return dongle.Verify.FromRawString(hexPwd, pwd).ByBcrypt().ToBool()
}
