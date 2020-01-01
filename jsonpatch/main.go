package main

import (
	"encoding/json"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
)

// json官方指定的patch结构为 `{"op":"add|remove|replace", "path":"/a/b", "value":xxx}`
// 还原成golang struc就如下所示了。
type Patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

//需要针对，针对新增资源的操作只能在最外层整体进行，而不能针对其成员
//注意: 只是针对新增资源, 不知道是不是包的原因

func main() {
	original := `{"name":"woniu","ages":[100,200],"infos":[{"shortname":"st1","place":"kaifeng"}]}`

	p1 := Patch{
		Op:    "add",
		Path:  "/infos/1",
		Value: `{"shortname":"st0","place":"hefei"}`,
	}

	/*
	/infos/0是已有资源，不是新增资源。所以此包可以对其成员变量进行操作
	*/
	p2 := Patch{
		Op:    "replace",
		Path:  "/infos/0/shortname",
		Value: `wulala`,
	}

	/*
	/infos/1是新增资源整体Op(remove)，所以可以. 但是/infors/1/shortname
	的remove则就不行了，因为shortname是新增资源的里层成员变量了
	*/
	p3 := Patch{
		Op:    "remove",
		Path:  "/infos/1",
		Value: nil,
	}

	bb, _ := json.Marshal([]Patch{p1, p2, p3})
	patch, _ := jsonpatch.DecodePatch(bb)

	modified, err := patch.Apply([]byte(original))
	fmt.Println(string(modified), err)
}
