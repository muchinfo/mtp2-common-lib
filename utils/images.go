package utils

import (
	"encoding/base64"
	"fmt"
	"os"
)

func Base64SaveImageFile(base64Str string, folder string, extName string) (fileName string, err error) {
	// 解码 Base64
	imgBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		fmt.Printf("解码 Base64 发生错误: %v\n", err)
		return
	}

	// 生成文件名
	fileName = GetUUID() + "." + extName

	// 写入文件
	if err = os.WriteFile(folder+fileName, imgBytes, 0644); err != nil {
		fmt.Printf("写入文件发生错误: %v\n", err)
		return
	}

	fmt.Printf("文件写入成功: %s\n", folder+fileName)

	return
}

func Base64SaveFile(base64str string, folder string) (fileName string, err error) {
	// 生成文件名
	fileName = GetUUID()

	// 写入文件
	if err = os.WriteFile(folder+fileName, []byte(base64str), 0644); err != nil {
		fmt.Printf("写入文件发生错误: %v\n", err)
		return
	}

	fmt.Printf("文件写入成功: %s\n", folder+fileName)

	return
}
