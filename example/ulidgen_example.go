package main

import (
	"fmt"

	"github.com/muchinfo/mtp2-common-lib/ulidgen"
)

func RunULIDGenExample() {
	// 生成标准 ULID
	id, err := ulidgen.GenerateULID()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("标准ULID:", id)

	// 生成18位短ULID
	id18, _ := ulidgen.GenerateShortULID(18)
	fmt.Println("18位ULID:", id18)

	// 生成带前缀且总长20位的ULID
	id20, _ := ulidgen.GenerateULIDWithPrefix("JYI", 20)
	fmt.Println("带前缀20位ULID:", id20)
}
