package main

import (
	"fmt"
	"regexp"
)

func main() {
	/**
	测试模式是否匹配字符串
	括号里面的意思是：至少有一个 a-z 之间的字符存在
	*/
	match, _ := regexp.MatchString("p([a-z]+)ch", "peach")
	fmt.Println(match)

	// 使用 Compile 来使用一个优化过的正则对象
	r, _ := regexp.Compile("p([a-z]+)ch")
	fmt.Println(r.MatchString("peach"))

	// 检测字符串参数中是否存在正则所约束的匹配
	fmt.Println(r.FindString("peach punch"))

}
