package consts

// 算法类型
type AlgorithmType struct {
	Name  string `json:"name"`  // 算法名称
	Label string `json:"label"` // 算法标签
}

// 检查算法类型是否存在
func CheckAlgorithmTypeValidity(algTypeName string) bool {
	algTypeValid := false
	for _, algType := range AlgorithmTypeList {
		if algType.Name == algTypeName {
			algTypeValid = true
			break
		}
	}

	return algTypeValid
}

// 算法类型列表 暂时写死固定这样 TODO: 后期可能从数据库动态获取
var AlgorithmTypeList = []AlgorithmType{
	{Name: "obj_detect", Label: "对象检测"},
	{Name: "obj_seg", Label: "对象分割"},
}
