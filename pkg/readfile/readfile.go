package readfile

import (
	"bufio"
	"os"
)

func Read(fileName string) (result []string, err error) {
	fp, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}
	return
}
