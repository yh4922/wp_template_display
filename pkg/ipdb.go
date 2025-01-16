package pkg

import (
	_ "embed"

	"github.com/ipipdotnet/ipdb-go"
)

//go:embed libs/ip.ipdb
var ipdbBytes []byte

var ipLib *ipdb.City

func init() {
	db, err := ipdb.NewCityFromBytes(ipdbBytes)
	if err == nil {
		ipLib = db
	}
}

// 获取ip信息
func IplibGetInfo(ip string) string {
	info, err := ipLib.FindInfo(ip, "CN")
	if err != nil {
		return ""
	}

	// 如果ip是本机地址，则返回本机地址
	if info.CountryName == "本机地址" {
		return "本机地址"
	}

	if info.CountryName == "局域网" {
		return "局域网"
	}

	// 国家 省份 城市
	return info.CountryName + " " + info.RegionName + " " + info.CityName
}
