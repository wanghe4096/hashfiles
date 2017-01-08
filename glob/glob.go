/* =============================================================================
glob 包实现了带有通配符（*、?）的字符串的匹配算法。

用法1（先对模式字符串进行编译，再进行匹配。此方法代码繁琐，但批量匹配时，效率高）：

1、创建 glob 对象（需要提供模式字符串）：
g := glob.Compile(`*???.txt`) // 匹配“文件名至少包含 3 个字符（可以是汉字）的文本文件”

2、设置相关选项：
g.CaseMind = true // 是否区分大小写（默认不区分大小写）

3、开始匹配：
ok := g.Match([]byte("somefile.txt")) // 匹配字节列表，成功返回 true，失败返回 false
ok := g.MatchString("somefile.txt")   // 匹配字符串，成功返回 true，失败返回 false

用法2（直接进行匹配，此方法代码简洁，但批量匹配时，效率低）：

// 匹配字节列表，成功返回 true，失败返回 false（最后一个参数表示“是否区分大小写”）
ok := glob.Match(`*???.txt`, []byte("somefile.txt"), true)

// 匹配字符串，成功返回 true，匹配失败返回 false（最后一个参数表示“是否区分大小写”）
ok := glob.MatchString(`*???.txt`, "somefile.txt", false)

注意：模式中原始的 *、?、\ 字符要经过转义（\*、\?、\\）才能代表字符本身，
　　　如果 \ 不在 *、?、\ 之前，可以不用转义。
============================================================================= */

package glob

import (
	"unicode/utf8"
)

// 主结构体（用来完成通配符匹配）
type Glob struct {
	CaseMind bool         // 是否区分大小写
	subPatts []subPattern // 子模式列表
}

// 用来存放子模式的结构体
type subPattern struct {
	patt     []byte          // 子模式字符串
	subExprs []subExpression // 子表达式列表（每个子模式都包含子表达式列表）
	mode     int             // 子模式的匹配方式
}

// 用来存放子表达式的结构体
type subExpression struct {
	expr []byte // 子表达式字符串
	mode int    // 子表达式的匹配方式
}

// 子模式或子表达式的匹配方式常量
const (
	matchStart = iota // 匹配 str 起始位置
	matchAny          // 匹配 str 任意位置
	matchEnd          // 匹配 str 结束位置
	matchFull         // 与 str 完整匹配
	matchOne          // 匹配 str 第一个字符（? 的情况）
)

// 编译“模式字符串”，得到 Glob 对象
func Compile(pattern string) *Glob {
	g := &Glob{
		CaseMind: false, // 默认不区分大小写
	}
	g.compile([]byte(pattern)) // 编译“模式字符串”的实现过程
	return g
}

// 编译“模式字符串”，获取子模式列表，和子表达式列表（子表达式由 getSubExprs 处理）
func (g *Glob) compile(pattern []byte) {
	var escape bool           // 用于标记是否处于转义状态
	var subPatt subPattern    // 当前正在处理的子模式
	subPatt.mode = matchStart // 默认匹配方式：匹配起始位置
	// 寻找 * 字符
	for _, c := range pattern {
		switch c {
		// 如果 c 为特殊字符
		case '*', '?', '\\':
			if escape { // 如果当前处于转义状态，则取消转义状态
				subPatt.patt = append(subPatt.patt, c)
				escape = false
			} else { // 如果当前处于非转义状态，则处理特殊字符
				switch c {
				case '\\': // 遇到 \ 则进入转义状态，继续处理
					// 保留转义字符，交给 getSubExprs 处理
					subPatt.patt = append(subPatt.patt, c)
					escape = true
				case '*':
					// 遇到 * 字符，处理 * 之前的内容
					if subPatt.patt != nil {
						// 对 subPatt.patt 中的内容进行分析，得到 subPatt.subExprs
						subPatt.subExprs = g.getSubExprs(subPatt.patt)
						// 将 subPatt 写入 g.subPatts（至此，一个子模式分析完毕）
						g.subPatts = append(g.subPatts, subPatt)
						// 复位 subPatt 元素
						subPatt.patt = nil
						subPatt.subExprs = nil
					}
					// 然后标记 * 之后的内容为 matchAny 模式，继续处理 * 之后的内容
					subPatt.mode = matchAny
				case '?':
					subPatt.patt = append(subPatt.patt, c) // 直接写入 ? 字符
				}
			}
		// 如果 c 为普通字符，则直接写入
		default:
			subPatt.patt = append(subPatt.patt, c)
		}
	}

	if len(subPatt.patt) > 0 {
		// 处理以普通字符结尾的情况
		switch subPatt.mode {
		case matchStart:
			subPatt.mode = matchFull // 完全匹配
		case matchAny:
			subPatt.mode = matchEnd // 尾部匹配
		}
		// 对 subPatt.patt 中的内容进行分析，得到 subPatt.subExprs
		subPatt.subExprs = g.getSubExprs(subPatt.patt)
		// 将 subPatt 写入 g.subPatts（至此，最后一个子模式分析完毕）
		g.subPatts = append(g.subPatts, subPatt)
	}
	// 到此，除了上面两种情况外，剩下的就是 subPatt.patt 为空的情况了：
	// subPatt.patt 为空表示整个模式以 * 结尾，此时只要 * 之前的子模式匹配成功，
	// 则整个模式就匹配成功了，所以这里就忽略了尾部 * 字符的匹配。
}

// 获取子表达式列表
func (g *Glob) getSubExprs(subPatt []byte) (result []subExpression) {
	var escape bool           // 用于标记是否处于转义状态
	var subExpr subExpression // 用于存储当前找到的子模式
	subExpr.mode = matchStart // 默认匹配方式
	// 寻找 ? 字符
	for _, c := range subPatt {
		switch c {
		// 如果 c 为特殊字符
		case '*', '?', '\\':
			if escape { // 如果当前处于转义状态，则取消转义状态，并写入转义后的字符
				subExpr.expr = append(subExpr.expr, c)
				escape = false
			} else { // 如果当前处于非转义状态，则处理特殊字符
				switch c {
				case '\\': // 遇到 \ 则进入转义状态，继续处理
					escape = true
				case '?':
					// 遇到 ?，将 ? 之前的内容写入 subExpr.expr
					if subExpr.expr != nil {
						result = append(result, subExpr)
						subExpr.expr = nil
					}
					// 将 ? 单独写入 subExpr
					subExpr.mode = matchOne
					result = append(result, subExpr)
					// 恢复默认匹配方式
					subExpr.mode = matchStart
					// 此处不允许出现 case '*': 的情况，因为单独的 * 字符，
					// 在 g.compile 中已经处理过了
					// case '*':
				}
			}
		// 如果 c 为普通字符
		default:
			if escape {
				// 处理 \ 后面跟的不是 \ 或 * 或 ? 的情况，写入之前的转义符号 \
				subExpr.expr = append(subExpr.expr, '\\')
				escape = false
			}
			// 写入当前字符
			subExpr.expr = append(subExpr.expr, c)
		}
	}
	if escape {
		// 处理整个子模式以 \ 结尾的情况（将尾部的 \ 当作普通字符处理）
		subExpr.expr = append(subExpr.expr, '\\')
		escape = false
	}
	if len(subExpr.expr) > 0 {
		// 处理最后的普通字符
		result = append(result, subExpr)
	}
	return result
}

// 判断 pattern 和 b 是否匹配
func (g *Glob) Match(b []byte) (matched bool) {
	matched = g.match(b)
	return
}

// 判断 pattern 和 s 是否匹配
func (g *Glob) MatchString(s string) (matched bool) {
	matched = g.match([]byte(s))
	return
}

// 通配符匹配过程的实现代码
func (g *Glob) match(b []byte) (matched bool) {
	// 循环获取子模式
	for _, subPatt := range g.subPatts {
		// 根据匹配方式进行判断
		switch subPatt.mode {
		case matchStart:
			if ok, r := g.hasPrefix(b, subPatt.subExprs); ok {
				// 匹配成功，去掉 b 中匹配的部分，继续处理剩下的部分
				b = r
			} else {
				return false
			}
		case matchAny:
			if ok, r := g.hasMidfix(b, subPatt.subExprs); ok {
				// 匹配成功，去掉 b 中匹配的部分，继续处理剩下的部分
				b = r
			} else {
				return false
			}
		case matchEnd:
			// 最后一个匹配项，直接返回匹配结果
			return g.hasSuffix(b, subPatt.subExprs)
		case matchFull:
			// 最后一个匹配项，直接返回匹配结果
			return g.same(b, subPatt.subExprs)
		}
	}
	// 所有子模式都匹配成功，返回 true
	return true
}

// 判断两个字节序列是否相同
// a,b：要比较的两个字节序列
// CaseMind：是否区分大小写
func equal(a, b []byte, CaseMind bool) bool {
	// 长度判断
	if len(a) != len(b) {
		return false
	}
	if CaseMind {
		for i, c := range a {
			// 直接比较
			if c != b[i] {
				return false
			}
		}
	} else {
		for i, c := range a {
			// 先转换大小写
			switch {
			case c > b[i] && c >= 'a' && c <= 'z':
				c = c - 'a' + 'A'
			case c < b[i] && c >= 'A' && c <= 'Z':
				c = c - 'A' + 'a'
			}
			// 再进行比较
			if c != b[i] {
				return false
			}
		}
	}
	// 全部都匹配，返回 true
	return true
}

// 判断带 ? 的子模式 subExprs 是否在 b 的开头位置
// 如果匹配成功，则返回 true 和去掉匹配部分之后的 b，否则返回 false, nil
func (g *Glob) hasPrefix(b []byte, subExprs []subExpression) (bool, []byte) {
	// 遍历子串列表
	for _, sub := range subExprs {
		switch sub.mode {
		case matchStart: // 如果是普通字符
			// 长度不够，或者不匹配
			if len(b) < len(sub.expr) ||
				!equal(b[:len(sub.expr)], sub.expr, g.CaseMind) {
				return false, nil
			}
			// 匹配成功，去掉 b 中匹配的部分，继续处理剩下的部分
			b = b[len(sub.expr):]
		case matchOne: // 如果是 ? 字符
			// 长度不够，返回 false
			if len(b) == 0 {
				return false, nil
			}
			// 长度够，去掉 b 中匹配的部分，继续处理剩下的部分
			_, size := utf8.DecodeRune(b)
			b = b[size:]
		}
	}
	// 所有子串都匹配成功，返回 true
	return true, b
}

// 判断带 ? 的子模式 subExprs 是否在 b 的中间位置
// 如果匹配成功，则返回 true 和去掉匹配部分之后的 b，否则返回 false, nil
func (g *Glob) hasMidfix(b []byte, subExprs []subExpression) (bool, []byte) {
	for _, sub := range subExprs {
		switch sub.mode {
		case matchStart: // 如果是普通字符
			for len(b) > 0 {
				// 找出 s 中与 sub.expr[0] 匹配的位置
				if b[0] == sub.expr[0] {
					// 长度不够，返回 false
					if len(b) < len(sub.expr) {
						return false, nil
					}
					// 长度够，截取 b 进行比较
					if equal(b[:len(sub.expr)], sub.expr, g.CaseMind) {
						// 匹配成功，去掉 b 中匹配的部分，继续处理剩下的部分
						b = b[len(sub.expr):]
						break
					}
				}
				// 继续比较下一个首字符
				b = b[1:]
			}
		case matchOne: // 如果是 ? 字符
			// 长度不够，返回 false
			if len(b) == 0 {
				return false, nil
			}
			// 长度够，去掉 b 中匹配的部分，继续处理剩下的部分
			_, size := utf8.DecodeRune(b)
			b = b[size:]
		}
	}
	// 所有子串都匹配成功，返回 true
	return true, b
}

// 判断带 ? 的子模式 subExprs 是否在 b 的结束位置
// 如果匹配成功，则返回 true，否则返回 false
func (g *Glob) hasSuffix(b []byte, subExprs []subExpression) bool {
	// 遍历子串列表，这里要倒序处理，所以不能用 for range 方式
	for i := len(subExprs) - 1; i >= 0; i-- {
		// 获取当前子表达式
		sub := subExprs[i]
		switch sub.mode {
		case matchStart: // 如果是普通字符
			// 长度不够，或者不匹配
			if len(b) < len(sub.expr) ||
				!equal(b[len(b)-len(sub.expr):], sub.expr, g.CaseMind) {
				return false
			}
			// 匹配成功，去掉 b 中匹配的部分，继续处理剩下的部分
			b = b[:len(b)-len(sub.expr)]
		case matchOne: // 如果是 ? 字符
			// 长度不够
			if len(b) == 0 {
				return false
			}
			// 长度够，去掉 b 中匹配的部分，继续处理剩下的部分
			_, size := utf8.DecodeRune(b)
			b = b[:len(b)-size]
			continue
		}
	}
	// 所有子串都匹配成功，返回 true
	return true
}

// 判断带 ? 的子模式 subExprs 是否与 b 完全匹配
// 如果匹配成功，则返回 true，否则返回 false
func (g *Glob) same(b []byte, subExprs []subExpression) bool {
	// 头部匹配成后，不再有剩余的 b 字符串，则表示完整匹配
	ok, r := g.hasPrefix(b, subExprs)
	return ok && len(r) == 0
}

// 判断 pattern 和 b 是否匹配，casemind 表示是否区分大小写
func Match(pattern string, b []byte, casemind bool) (matched bool) {
	g := Compile(pattern)
	g.CaseMind = casemind
	return g.Match(b)
}

// 判断 pattern 和 s 是否匹配，casemind 表示是否区分大小写
func MatchString(pattern, s string, casemind bool) (matched bool) {
	g := Compile(pattern)
	g.CaseMind = casemind
	return g.MatchString(s)
}
