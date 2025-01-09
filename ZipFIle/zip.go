package zipfile

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jhillyerd/enmime"
)

// 遍历文件夹并解析文本文件
func ParseTextFilesInFolder(folderPath string) error {
	return filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查是否为文件
		if !info.IsDir() {
			// 检查文件扩展名是否为文本文件
			// 读取文件内容
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			// 使用enmime解析文件内容
			env, err := enmime.ReadEnvelope(strings.NewReader(string(content)))
			if err != nil {
				return err
			}

			// 在这里处理解析后的邮件内容
			// 例如：打印主题和正文
			println("文件路径:", path)
			println("主题:", env.GetHeader("Subject"))
			println("正文Text:", env.Text)
			println("正文HTML:", env.HTML)
			println("--------------------------------------------------------------------------------------------------------------------------------------------------")
		}

		return nil
	})
}
