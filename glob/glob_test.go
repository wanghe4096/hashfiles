package glob

import (
	"fmt"
	"testing"
)

// 测试 compile（去掉函数名前面的 x 字符，可以执行测试）
func xTestCompile(t *testing.T) {
	tcompile(`a******b`, t)
	tcompile(`a*?*?*?b`, t)
	tcompile(`a?a*b?b*c?c`, t)
	tcompile(`a?a*b?b*c?c*`, t)
	tcompile(`*a?a*b?b*c?c`, t)
	tcompile(`*a?a*b?b*c?c*`, t)
}

// 辅助测试函数
func tcompile(str string, t *testing.T) {
	g := Compile(str)
	fmt.Println(showSubPatt(g.subPatts))
}

// 辅助测试函数，用于输出 subPattern 序列
func showSubPatt(subPatts []subPattern) (result string) {
	for _, sub := range subPatts {
		switch sub.mode {
		case matchStart:
			result += fmt.Sprintf(`<%s,首>`, sub.patt)
		case matchAny:
			result += fmt.Sprintf(`<%s,中>`, sub.patt)
		case matchEnd:
			result += fmt.Sprintf(`<%s,尾>`, sub.patt)
		case matchFull:
			result += fmt.Sprintf(`<%s,全>`, sub.patt)
		}
	}
	return
}

// 测试 g.getSubExprs（去掉函数名前面的 x 字符，可以执行测试）
func xTestGetSubExprs(t *testing.T) {
	tgetSubExprs(`a\`, t)
	tgetSubExprs(`\a`, t)
	tgetSubExprs(`a\\\*\?`, t)
	tgetSubExprs(`a?a`, t)
	tgetSubExprs(`?a?`, t)
	tgetSubExprs(`a??`, t)
	tgetSubExprs(`??a`, t)
}

// 辅助测试函数
func tgetSubExprs(str string, t *testing.T) {
	g := Compile("")
	fmt.Println(showSubExpr(g.getSubExprs([]byte(str))))
}

// 辅助测试函数，用于输出 subExpression 序列
func showSubExpr(subExprs []subExpression) (result string) {
	for _, sub := range subExprs {
		switch sub.mode {
		case matchStart:
			result += fmt.Sprintf(`<%s,首>`, sub.expr)
		case matchOne:
			result += fmt.Sprintf(`<单>`)
		}
	}
	return
}

// 测试 g.hasPrefix（去掉函数名前面的 x 字符，可以执行测试）
func xTestHasPrefix(t *testing.T) {
	thasPrefix(`AbAcc`, `aba`, t) // 区分大小写测试
	thasPrefix(`abacc`, `a?a`, t)
	thasPrefix(`babcc`, `?a?`, t)
	thasPrefix(`abbcc`, `a??`, t)
	thasPrefix(`bbacc`, `??a`, t)
}

// 辅助测试函数
func thasPrefix(str, substr string, t *testing.T) {
	g := Compile("")
	g.CaseMind = false
	ok, rs := g.hasPrefix([]byte(str), g.getSubExprs([]byte(substr)))
	fmt.Printf("%s 匹配 %s = %v 剩余 %s\n", neat(str), neat(substr), ok, rs)
}

// 测试 g.hasMidfix（去掉函数名前面的 x 字符，可以执行测试）
func xTestHasMidfix(t *testing.T) {
	thasMidfix(`ccAbAcc`, `aba`, t) // 区分大小写测试
	thasMidfix(`ccabacc`, `a?a`, t)
	thasMidfix(`ccbabcc`, `?a?`, t)
	thasMidfix(`ccabbcc`, `a??`, t)
	thasMidfix(`ccbbacc`, `??a`, t)
}

// 辅助测试函数
func thasMidfix(str, substr string, t *testing.T) {
	g := Compile("")
	g.CaseMind = false
	ok, rs := g.hasMidfix([]byte(str), g.getSubExprs([]byte(substr)))
	fmt.Printf("%s 匹配 %s = %v 剩余 %s\n", neat(str), neat(substr), ok, rs)
}

// 测试 g.hasSuffix（去掉函数名前面的 x 字符，可以执行测试）
func xTestHasSuffix(t *testing.T) {
	thasSuffix(`ccAbA`, `aba`, t) // 区分大小写测试
	thasSuffix(`ccaba`, `a?a`, t)
	thasSuffix(`ccbab`, `?a?`, t)
	thasSuffix(`ccabb`, `a??`, t)
	thasSuffix(`ccbba`, `??a`, t)
}

// 辅助测试函数
func thasSuffix(str, substr string, t *testing.T) {
	g := Compile("")
	g.CaseMind = false
	ok := g.hasSuffix([]byte(str), g.getSubExprs([]byte(substr)))
	fmt.Printf("%s 匹配 %s = %v\n", neat(str), neat(substr), ok)
}

// 测试 g.same（去掉函数名前面的 x 字符，可以执行测试）
func xTestSame(t *testing.T) {
	tsame(`AbA`, `aba`, t) // 区分大小写测试
	tsame(`aba`, `a?a`, t)
	tsame(`bab`, `?a?`, t)
	tsame(`abb`, `a??`, t)
	tsame(`bba`, `??a`, t)
}

// 辅助测试函数
func tsame(str, substr string, t *testing.T) {
	g := Compile("")
	g.CaseMind = false
	ok := g.same([]byte(str), g.getSubExprs([]byte(substr)))
	fmt.Printf("%s 匹配 %s = %v\n", neat(str), neat(substr), ok)
}

// 测试 g.Match（去掉函数名前面的 x 字符，可以执行测试）
func TestMatch(t *testing.T) {
	tMatch(`a*`, `abc`, t)
	tMatch(`*a`, `cba`, t)
	tMatch(`a?`, `ab`, t)
	tMatch(`?a`, `ba`, t)
	tMatch(`a\`, `a\`, t)
	tMatch(`\a`, `\a`, t)
	tMatch(`a\*\?\\`, `a*?\`, t)
	tMatch(`a*b?b`, `aaaabbbb`, t)
	tMatch(`a**b`, `aaaabbbb`, t)
	tMatch(`a*?b`, `aaaabbbb`, t)
	tMatch(`a??b`, `abab`, t)
	tMatch(`a*?**?*?b`, `aaaabbbb`, t)
	tMatch(`A*?ab?*?b`, `aaaabbbB`, t) // 区分大小写测试
}

// 辅助测试函数
func tMatch(pattern, str string, t *testing.T) {
	g := Compile(pattern)
	g.CaseMind = true // 这个测试区分大小写
	ok := g.Match([]byte(str))
	fmt.Printf("%s 匹配 %s = %v\n", neat(str), neat(pattern), ok)
}

// 将字符串扩展到指定长度
func neat(s string) string {
	// 统一显示宽度为 10 个字节
	const maxLen = 10
	for len(s) < maxLen {
		s += ` `
	}
	return s
}
