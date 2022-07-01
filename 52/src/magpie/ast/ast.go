package ast

import (
	"bytes"
	"fmt"
	"magpie/token"
	"strings"
	"unicode/utf8"
)

type Node interface {
	Pos() token.Position // position of first character belonging to the node
	End() token.Position // position of first character immediately after the node

	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
	Imports    map[string]*ImportStatement
}

func (p *Program) Pos() token.Position {
	if len(p.Statements) > 0 {
		return p.Statements[0].Pos()
	}
	return token.Position{}
}

func (p *Program) End() token.Position {
	aLen := len(p.Statements)
	if aLen > 0 {
		return p.Statements[aLen-1].End()
	}
	return token.Position{}
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type ImportStatement struct {
	Token      token.Token
	ImportPath string
	Program    *Program
}

func (is *ImportStatement) Pos() token.Position {
	return is.Token.Pos
}

func (is *ImportStatement) End() token.Position {
	length := utf8.RuneCountInString(is.ImportPath)
	return token.Position{Filename: is.Token.Pos.Filename, Line: is.Token.Pos.Line, Col: is.Token.Pos.Col + length}
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string {
	var out bytes.Buffer

	out.WriteString(is.TokenLiteral())
	out.WriteString(" ")
	out.WriteString(is.ImportPath)

	return out.String()
}

//let <identifier1>,<identifier2>,... = <expression1>,<expression2>,...
type LetStatement struct {
	Token  token.Token
	Names  []*Identifier
	Values []Expression
}

func (ls *LetStatement) Pos() token.Position {
	return ls.Token.Pos
}

func (ls *LetStatement) End() token.Position {
	aLen := len(ls.Values)
	if aLen > 0 {
		return ls.Values[aLen-1].End()
	}

	return ls.Names[0].End()
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")

	names := []string{}
	for _, name := range ls.Names {
		names = append(names, name.String())
	}
	out.WriteString(strings.Join(names, ", "))

	if len(ls.Values) == 0 { //e.g. 'let x'
		out.WriteString(";")
		return out.String()
	}

	out.WriteString(" = ")

	values := []string{}
	for _, value := range ls.Values {
		values = append(values, value.String())
	}
	out.WriteString(strings.Join(values, ", "))

	return out.String()
}

type ReturnStatement struct {
	Token        token.Token // the 'return' token
	ReturnValue  Expression  //for old campatibility
	ReturnValues []Expression
}

func (rs *ReturnStatement) Pos() token.Position {
	return rs.Token.Pos
}

func (rs *ReturnStatement) End() token.Position {
	aLen := len(rs.ReturnValues)
	if aLen > 0 {
		return rs.ReturnValues[aLen-1].End()
	}

	return token.Position{Filename: rs.Token.Pos.Filename, Line: rs.Token.Pos.Line, Col: rs.Token.Pos.Col + len(rs.Token.Literal)}

}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	//	if rs.ReturnValue != nil {
	//		out.WriteString(rs.ReturnValue.String())
	//	}

	values := []string{}
	for _, value := range rs.ReturnValues {
		values = append(values, value.String())
	}
	out.WriteString(strings.Join(values, ", "))

	out.WriteString(";")

	return out.String()
}

//tailcall funcCall(param1, param2, ...)
type TailCallStatement struct {
	Token token.Token // the 'tailcall' token
	Call  Expression
}

func (ts *TailCallStatement) Pos() token.Position {
	return ts.Token.Pos
}

func (ts *TailCallStatement) End() token.Position {
	return ts.Call.End()
}

func (ts *TailCallStatement) statementNode()       {}
func (ts *TailCallStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *TailCallStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ts.TokenLiteral() + " ")
	out.WriteString(ts.Call.String())
	out.WriteString(";")

	return out.String()
}

type BlockStatement struct {
	Token       token.Token
	Statements  []Statement
	RBraceToken token.Token //used in End() method
}

func (bs *BlockStatement) Pos() token.Position {
	return bs.Token.Pos

}

func (bs *BlockStatement) End() token.Position {
	return token.Position{Filename: bs.Token.Pos.Filename, Line: bs.RBraceToken.Pos.Line, Col: bs.RBraceToken.Pos.Col + 1}
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		str := s.String()

		out.WriteString(str)
		if str[len(str)-1:] != ";" {
			out.WriteString(";")
		}
	}

	return out.String()
}

type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) Pos() token.Position {
	return es.Token.Pos
}

func (es *ExpressionStatement) End() token.Position {
	return es.Expression.End()
}
func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// 1 + 2 * 3
type InfixExpression struct {
	Token        token.Token
	Operator     string
	Right        Expression
	Left         Expression
	HasNext      bool
	NextOperator string
	Next         Expression
}

func (ie *InfixExpression) Pos() token.Position { return ie.Token.Pos }
func (ie *InfixExpression) End() token.Position { return ie.Right.End() }

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())

	if ie.HasNext {
		out.WriteString(" " + ie.NextOperator + " ")
		out.WriteString(ie.Next.String())
	}

	out.WriteString(")")

	return out.String()
}

// -2, -3
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) Pos() token.Position { return pe.Token.Pos }
func (pe *PrefixExpression) End() token.Position { return pe.Right.End() }

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type PostfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
}

func (pe *PostfixExpression) Pos() token.Position {
	return pe.Token.Pos
}

func (pe *PostfixExpression) End() token.Position {
	ret := pe.Left.End()
	ret.Col = ret.Col + len(pe.Operator)
	return ret
}

func (pe *PostfixExpression) expressionNode() {}

func (pe *PostfixExpression) TokenLiteral() string {
	return pe.Token.Literal
}

func (pe *PostfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Left.String())
	out.WriteString(pe.Operator)
	out.WriteString(")")

	return out.String()
}

type NumberLiteral struct {
	Token token.Token
	Value float64
}

func (nl *NumberLiteral) Pos() token.Position { return nl.Token.Pos }
func (nl *NumberLiteral) End() token.Position {
	length := utf8.RuneCountInString(nl.Token.Literal)
	pos := nl.Token.Pos
	return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
}

func (nl *NumberLiteral) expressionNode()      {}
func (nl *NumberLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NumberLiteral) String() string       { return nl.Token.Literal }

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) Pos() token.Position { return i.Token.Pos }
func (i *Identifier) End() token.Position {
	length := utf8.RuneCountInString(i.Value)
	return token.Position{Filename: i.Token.Pos.Filename, Line: i.Token.Pos.Line, Col: i.Token.Pos.Col + length}
}
func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

type NilLiteral struct {
	Token token.Token
}

func (n *NilLiteral) Pos() token.Position {
	return n.Token.Pos
}

func (n *NilLiteral) End() token.Position {
	length := len(n.Token.Literal)
	pos := n.Token.Pos
	return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
}

func (n *NilLiteral) expressionNode()      {}
func (n *NilLiteral) TokenLiteral() string { return n.Token.Literal }
func (n *NilLiteral) String() string       { return n.Token.Literal }

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (b *BooleanLiteral) Pos() token.Position {
	return b.Token.Pos
}

func (b *BooleanLiteral) End() token.Position {
	length := utf8.RuneCountInString(b.Token.Literal)
	pos := b.Token.Pos
	return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
}

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }
func (b *BooleanLiteral) String() string       { return b.Token.Literal }

type StringLiteral struct {
	Token token.Token
	Value string
}

func (s *StringLiteral) Pos() token.Position {
	return s.Token.Pos
}

func (s *StringLiteral) End() token.Position {
	length := utf8.RuneCountInString(s.Value)
	return token.Position{Filename: s.Token.Pos.Filename, Line: s.Token.Pos.Line, Col: s.Token.Pos.Col + length}
}

func (s *StringLiteral) expressionNode()      {}
func (s *StringLiteral) TokenLiteral() string { return s.Token.Literal }
func (s *StringLiteral) String() string       { return s.Token.Literal }

type FunctionLiteral struct {
	Token      token.Token // The 'fn' token
	Name       string      // function's name
	Parameters []*Identifier
	Variadic   bool
	Body       *BlockStatement
}

func (fl *FunctionLiteral) Pos() token.Position {
	return fl.Token.Pos
}

func (fl *FunctionLiteral) End() token.Position {
	return fl.Body.End()
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	if fl.Name != "" {
		out.WriteString(" ")
		out.WriteString(fl.Name)
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	if fl.Variadic {
		out.WriteString("...")
	}
	out.WriteString(") {")
	out.WriteString(fl.Body.String())
	out.WriteString("}")

	return out.String()
}

type ArrayLiteral struct {
	Token   token.Token
	Members []Expression
}

func (a *ArrayLiteral) Pos() token.Position {
	return a.Token.Pos
}

func (a *ArrayLiteral) End() token.Position {
	aLen := len(a.Members)
	if aLen > 0 {
		return a.Members[aLen-1].End()
	}
	return a.Token.Pos
}

func (a *ArrayLiteral) expressionNode()      {}
func (a *ArrayLiteral) TokenLiteral() string { return a.Token.Literal }
func (a *ArrayLiteral) String() string {
	var out bytes.Buffer

	members := []string{}
	for _, m := range a.Members {
		members = append(members, m.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(members, ", "))
	out.WriteString("]")
	return out.String()
}

type TupleLiteral struct {
	Token   token.Token
	Members []Expression
}

func (t *TupleLiteral) Pos() token.Position {
	return t.Token.Pos
}

func (t *TupleLiteral) End() token.Position {
	tLen := len(t.Members)
	if tLen > 0 {
		return t.Members[tLen-1].End()
	}
	return t.Token.Pos
}

func (t *TupleLiteral) expressionNode()      {}
func (t *TupleLiteral) TokenLiteral() string { return t.Token.Literal }
func (t *TupleLiteral) String() string {
	var out bytes.Buffer

	out.WriteString("(")

	members := []string{}
	for _, m := range t.Members {
		members = append(members, m.String())
	}

	out.WriteString(strings.Join(members, ", "))
	out.WriteString(")")

	return out.String()
}

//<Left-Expression>[<Index-Expression>]
type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) Pos() token.Position {
	return ie.Token.Pos
}

func (ie *IndexExpression) End() token.Position {
	return ie.Index.End()
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("]")
	out.WriteString(")")
	return out.String()
}

type HashLiteral struct {
	Token       token.Token
	Pairs       map[Expression]Expression
	RBraceToken token.Token
	IsOrdered   bool
	Order       []Expression //For keeping the order of the hash key
}

func (h *HashLiteral) Pos() token.Position {
	return h.Token.Pos
}

func (h *HashLiteral) End() token.Position {
	return token.Position{Filename: h.Token.Pos.Filename, Line: h.RBraceToken.Pos.Line, Col: h.RBraceToken.Pos.Col + 1}
}

func (h *HashLiteral) expressionNode()      {}
func (h *HashLiteral) TokenLiteral() string { return h.Token.Literal }
func (h *HashLiteral) String() string {
	var out bytes.Buffer

	pairs := []string{}
	if h.IsOrdered {
		for _, key := range h.Order {
			value, _ := h.Pairs[key]
			pairs = append(pairs, key.String()+": "+value.String())
		}
	} else {
		for key, value := range h.Pairs {
			pairs = append(pairs, key.String()+":"+value.String())
		}
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
	Variadic  bool
}

func (ce *CallExpression) Pos() token.Position {
	length := utf8.RuneCountInString(ce.Function.String())
	return token.Position{Filename: ce.Token.Pos.Filename, Line: ce.Token.Pos.Line, Col: ce.Token.Pos.Col - length}
}

func (ce *CallExpression) End() token.Position {
	aLen := len(ce.Arguments)
	if aLen > 0 {
		return ce.Arguments[aLen-1].End()
	}
	return ce.Function.End()
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	if ce.Variadic {
		out.WriteString("...")
	}
	out.WriteString(")")
	return out.String()
}

type MethodCallExpression struct {
	Token  token.Token
	Object Expression
	Call   Expression
}

func (mc *MethodCallExpression) Pos() token.Position {
	return mc.Token.Pos
}

func (mc *MethodCallExpression) End() token.Position {
	return mc.Call.End()
}

func (mc *MethodCallExpression) expressionNode()      {}
func (mc *MethodCallExpression) TokenLiteral() string { return mc.Token.Literal }
func (mc *MethodCallExpression) String() string {
	var out bytes.Buffer
	out.WriteString(mc.Object.String())
	out.WriteString(".")
	out.WriteString(mc.Call.String())

	return out.String()
}

type IfExpression struct {
	Token       token.Token
	Conditions  []*IfConditionExpr //if or else-if part
	Alternative *BlockStatement    //else part
}

func (ifex *IfExpression) Pos() token.Position {
	return ifex.Token.Pos
}

func (ifex *IfExpression) End() token.Position {
	if ifex.Alternative != nil {
		return ifex.Alternative.End()
	}

	aLen := len(ifex.Conditions)
	return ifex.Conditions[aLen-1].End()
}

func (ifex *IfExpression) expressionNode()      {}
func (ifex *IfExpression) TokenLiteral() string { return ifex.Token.Literal }

func (ifex *IfExpression) String() string {
	var out bytes.Buffer

	for i, c := range ifex.Conditions {
		if i == 0 {
			out.WriteString("if ")
		} else {
			out.WriteString("elif ")
		}
		out.WriteString(c.String())
	}

	if ifex.Alternative != nil {
		out.WriteString(" else ")
		out.WriteString(" { ")
		out.WriteString(ifex.Alternative.String())
		out.WriteString(" }")
	}

	return out.String()
}

//if/else-if condition
type IfConditionExpr struct {
	Token token.Token
	Cond  Expression      //condition
	Body  *BlockStatement //body
}

func (ic *IfConditionExpr) Pos() token.Position {
	return ic.Token.Pos
}

func (ic *IfConditionExpr) End() token.Position {
	return ic.Body.End()
}

func (ic *IfConditionExpr) expressionNode()      {}
func (ic *IfConditionExpr) TokenLiteral() string { return ic.Token.Literal }

func (ic *IfConditionExpr) String() string {
	var out bytes.Buffer

	out.WriteString(ic.Cond.String())
	out.WriteString(" { ")
	out.WriteString(ic.Body.String())
	out.WriteString(" }")

	return out.String()
}

type MultiAssignStatement struct {
	Token  token.Token
	Names  []Expression
	Values []Expression
}

func (as *MultiAssignStatement) Pos() token.Position {
	return as.Token.Pos
}

func (as *MultiAssignStatement) End() token.Position {
	aLen := len(as.Values)
	if aLen > 0 {
		return as.Values[aLen-1].End()
	}

	return as.Values[0].End()
}

func (as *MultiAssignStatement) statementNode()       {}
func (as *MultiAssignStatement) TokenLiteral() string { return as.Token.Literal }

func (as *MultiAssignStatement) String() string {
	var out bytes.Buffer

	names := []string{}
	for _, name := range as.Names {
		names = append(names, name.String())
	}
	out.WriteString(strings.Join(names, ", "))

	out.WriteString(" = ")

	values := []string{}
	for _, value := range as.Values {
		values = append(values, value.String())
	}
	out.WriteString(strings.Join(values, ", "))

	return out.String()
}

type AssignExpression struct {
	Token token.Token
	Name  Expression
	Value Expression
}

func (ae *AssignExpression) Pos() token.Position {
	//return ae.Token.Pos
	return ae.Name.Pos()
}

func (ae *AssignExpression) End() token.Position {
	return ae.Value.End()
}

func (ae *AssignExpression) expressionNode()      {}
func (ae *AssignExpression) TokenLiteral() string { return ae.Token.Literal }

func (ae *AssignExpression) String() string {
	var out bytes.Buffer

	out.WriteString(ae.Name.String())
	//out.WriteString(" = ")
	out.WriteString(ae.Token.Literal)
	out.WriteString(ae.Value.String())

	return out.String()
}

type BreakExpression struct {
	Token token.Token
}

func (be *BreakExpression) Pos() token.Position {
	return be.Token.Pos
}

func (be *BreakExpression) End() token.Position {
	length := utf8.RuneCountInString(be.Token.Literal)
	pos := be.Token.Pos
	return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
}

func (be *BreakExpression) expressionNode()      {}
func (be *BreakExpression) TokenLiteral() string { return be.Token.Literal }

func (be *BreakExpression) String() string { return be.Token.Literal }

///////////////////////////////////////////////////////////
//                         CONTINUE                      //
///////////////////////////////////////////////////////////
type ContinueExpression struct {
	Token token.Token
}

func (ce *ContinueExpression) Pos() token.Position {
	return ce.Token.Pos
}

func (ce *ContinueExpression) End() token.Position {
	length := utf8.RuneCountInString(ce.Token.Literal)
	pos := ce.Token.Pos
	return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
}

func (ce *ContinueExpression) expressionNode()      {}
func (ce *ContinueExpression) TokenLiteral() string { return ce.Token.Literal }

func (ce *ContinueExpression) String() string { return ce.Token.Literal }

//c language like for loop
type CForLoop struct {
	Token  token.Token
	Init   Expression
	Cond   Expression
	Update Expression
	Block  *BlockStatement
}

func (fl *CForLoop) Pos() token.Position {
	return fl.Token.Pos
}

func (fl *CForLoop) End() token.Position {
	return fl.Block.End()
}

func (fl *CForLoop) expressionNode()      {}
func (fl *CForLoop) TokenLiteral() string { return fl.Token.Literal }

func (fl *CForLoop) String() string {
	var out bytes.Buffer

	out.WriteString("for")
	out.WriteString(" ( ")

	if fl.Init != nil {
		out.WriteString(fl.Init.String())
	}
	out.WriteString(" ; ")

	if fl.Cond != nil {
		out.WriteString(fl.Cond.String())
	}
	out.WriteString(" ; ")

	if fl.Update != nil {
		out.WriteString(fl.Update.String())
	}
	out.WriteString(" ) ")
	out.WriteString(" { ")
	out.WriteString(fl.Block.String())
	out.WriteString(" }")

	return out.String()
}

//for var in value { block }
type ForEachArrayLoop struct {
	Token token.Token
	Var   string
	Value Expression //value to range over
	Block *BlockStatement
}

func (fal *ForEachArrayLoop) Pos() token.Position {
	return fal.Token.Pos
}

func (fal *ForEachArrayLoop) End() token.Position {
	return fal.Block.End()
}

func (fal *ForEachArrayLoop) expressionNode()      {}
func (fal *ForEachArrayLoop) TokenLiteral() string { return fal.Token.Literal }

func (fal *ForEachArrayLoop) String() string {
	var out bytes.Buffer

	out.WriteString("for ")
	out.WriteString(fal.Var)
	out.WriteString(" in ")
	out.WriteString(fal.Value.String())
	out.WriteString(" { ")
	out.WriteString(fal.Block.String())
	out.WriteString(" }")

	return out.String()
}

//for key, value in X { block }
type ForEachMapLoop struct {
	Token token.Token
	Key   string
	Value string
	X     Expression //value to range over
	Block *BlockStatement
}

func (fml *ForEachMapLoop) Pos() token.Position {
	return fml.Token.Pos
}

func (fml *ForEachMapLoop) End() token.Position {
	return fml.Block.End()
}

func (fml *ForEachMapLoop) expressionNode()      {}
func (fml *ForEachMapLoop) TokenLiteral() string { return fml.Token.Literal }

func (fml *ForEachMapLoop) String() string {
	var out bytes.Buffer

	out.WriteString("for ")
	out.WriteString(fml.Key + ", " + fml.Value)
	out.WriteString(" in ")
	out.WriteString(fml.X.String())
	out.WriteString(" { ")
	out.WriteString(fml.Block.String())
	out.WriteString(" }")

	return out.String()
}

//for { block }
type ForEverLoop struct {
	Token token.Token
	Block *BlockStatement
}

func (fel *ForEverLoop) Pos() token.Position {
	return fel.Token.Pos
}

func (fel *ForEverLoop) End() token.Position {
	return fel.Block.End()
}

func (fel *ForEverLoop) expressionNode()      {}
func (fel *ForEverLoop) TokenLiteral() string { return fel.Token.Literal }

func (fel *ForEverLoop) String() string {
	var out bytes.Buffer

	out.WriteString("for ")
	out.WriteString(" { ")
	out.WriteString(fel.Block.String())
	out.WriteString(" }")

	return out.String()
}

//while condition { block }
type WhileLoop struct {
	Token     token.Token
	Condition Expression
	Block     *BlockStatement
}

func (wl *WhileLoop) Pos() token.Position {
	return wl.Token.Pos
}

func (wl *WhileLoop) End() token.Position {
	return wl.Block.End()
}

func (wl *WhileLoop) expressionNode()      {}
func (wl *WhileLoop) TokenLiteral() string { return wl.Token.Literal }

func (wl *WhileLoop) String() string {
	var out bytes.Buffer

	out.WriteString("while")
	out.WriteString(wl.Condition.String())
	out.WriteString("{")
	out.WriteString(wl.Block.String())
	out.WriteString("}")

	return out.String()
}

//do { block }
type DoLoop struct {
	Token token.Token
	Block *BlockStatement
}

func (dl *DoLoop) Pos() token.Position {
	return dl.Token.Pos
}

func (dl *DoLoop) End() token.Position {
	return dl.Block.End()
}

func (dl *DoLoop) expressionNode()      {}
func (dl *DoLoop) TokenLiteral() string { return dl.Token.Literal }

func (dl *DoLoop) String() string {
	var out bytes.Buffer

	out.WriteString("do")
	out.WriteString(" { ")
	out.WriteString(dl.Block.String())
	out.WriteString(" }")
	return out.String()
}

type RegExLiteral struct {
	Token token.Token
	Value string // value of the regular expression
}

func (rel *RegExLiteral) Pos() token.Position {
	return rel.Token.Pos
}

func (rel *RegExLiteral) End() token.Position {
	length := utf8.RuneCountInString(rel.Value)
	pos := rel.Token.Pos
	return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
}

func (rel *RegExLiteral) expressionNode()      {}
func (rel *RegExLiteral) TokenLiteral() string { return rel.Token.Literal }
func (rel *RegExLiteral) String() string {
	reg := rel.Value
	begin := strings.Index(reg, "(?")
	if begin == -1 {
		return fmt.Sprintf("/%s/", reg)
	}
	end := strings.Index(reg, ")")
	val := reg[end+1:]
	flag := reg[begin+2 : end]
	return fmt.Sprintf("/%s/%s", val, flag)
}

type StructStatement struct {
	Token token.Token
	Name  string //struct's name

	Block       *BlockStatement //used in the String() method
	RBraceToken token.Token     //used in End() method
}

func (s *StructStatement) Pos() token.Position {
	return s.Token.Pos
}

func (s *StructStatement) End() token.Position {
	return s.RBraceToken.Pos
}

func (s *StructStatement) statementNode()       {}
func (s *StructStatement) TokenLiteral() string { return s.Token.Literal }
func (s *StructStatement) String() string {
	var out bytes.Buffer

	out.WriteString(s.Token.Literal + " ")
	out.WriteString(s.Name)

	out.WriteString("{ ")
	out.WriteString(s.Block.String())
	out.WriteString(" }")

	return out.String()
}

/*
    switch Expr {
    case expr1, expr2, ... { block1 }
    case expr3, expr4, ... { block2 }
    ...
    default { block }
}
*/
type SwitchExpression struct {
	Token       token.Token
	Expr        Expression
	Cases       []*CaseExpression
	RBraceToken token.Token //used in End() method
}

func (se *SwitchExpression) Pos() token.Position {
	return se.Token.Pos
}

func (se *SwitchExpression) End() token.Position {
	return se.RBraceToken.Pos
}

func (se *SwitchExpression) expressionNode()      {}
func (se *SwitchExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SwitchExpression) String() string {
	var out bytes.Buffer
	out.WriteString("switch ")
	out.WriteString(se.Expr.String())
	out.WriteString("{ ")

	for _, item := range se.Cases {
		if item != nil {
			out.WriteString(item.String())
		}
	}
	out.WriteString(" }")

	return out.String()
}

/*
   case expr1, expr2, ... { block }
   default                { block }

*/
type CaseExpression struct {
	Token       token.Token
	Default     bool //default case or not
	Exprs       []Expression
	Block       *BlockStatement
	RBraceToken token.Token //used in End() method
}

func (ce *CaseExpression) Pos() token.Position {
	return ce.Token.Pos
}

func (ce *CaseExpression) End() token.Position {
	return ce.RBraceToken.Pos
}

func (ce *CaseExpression) expressionNode()      {}
func (ce *CaseExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CaseExpression) String() string {
	var out bytes.Buffer

	if ce.Default {
		out.WriteString("default ")
	} else {
		out.WriteString("case ")

		exprs := []string{}
		for _, expr := range ce.Exprs {
			exprs = append(exprs, expr.String())
		}
		out.WriteString(strings.Join(exprs, ","))
	}
	out.WriteString(ce.Block.String())
	return out.String()
}

type FallthroughExpression struct {
	Token token.Token
}

//t: through
func (t *FallthroughExpression) Pos() token.Position {
	return t.Token.Pos
}

func (t *FallthroughExpression) End() token.Position {
	length := utf8.RuneCountInString(t.Token.Literal)
	pos := t.Token.Pos
	return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
}

func (t *FallthroughExpression) expressionNode()      {}
func (t *FallthroughExpression) TokenLiteral() string { return t.Token.Literal }

func (t *FallthroughExpression) String() string { return t.Token.Literal }

//TryStmt provide "try/catch/finally" statement.
/*
   try {block }
   catch e { block }
   finally {block }
*/

//TryStmt provide "try/catch/finally" statement.
type TryStmt struct {
	Token   token.Token
	Try     *BlockStatement
	Var     string
	Catch   *BlockStatement
	Finally *BlockStatement
}

func (t *TryStmt) Pos() token.Position {
	return t.Token.Pos
}

func (t *TryStmt) End() token.Position {
	if t.Finally != nil {
		return t.Finally.End()
	}

	if t.Catch != nil {
		return t.Catch.End()
	}

	return t.Try.End()
}

func (t *TryStmt) statementNode()       {}
func (t *TryStmt) TokenLiteral() string { return t.Token.Literal }

func (t *TryStmt) String() string {
	var out bytes.Buffer

	out.WriteString("try { ")
	out.WriteString(t.Try.String())
	out.WriteString(" }")

	if t.Catch != nil {
		if t.Var != "" {
			out.WriteString(" catch " + t.Var + " { ")
		} else {
			out.WriteString(" catch { ")
		}
		out.WriteString(t.Catch.String())
		out.WriteString(" }")
	}

	if t.Finally != nil {
		out.WriteString(" finally { ")
		out.WriteString(t.Finally.String())
		out.WriteString(" }")
	}

	return out.String()
}

//throw <expression>
type ThrowStmt struct {
	Token token.Token
	Expr  Expression
}

func (ts *ThrowStmt) Pos() token.Position {
	return ts.Token.Pos
}

func (ts *ThrowStmt) End() token.Position {
	return ts.Expr.End()
}

func (ts *ThrowStmt) statementNode()       {}
func (ts *ThrowStmt) TokenLiteral() string { return ts.Token.Literal }

func (ts *ThrowStmt) String() string {
	var out bytes.Buffer

	out.WriteString("throw ")
	out.WriteString(ts.Expr.String())
	out.WriteString(";")

	return out.String()
}

//@Func Decorated
//e.g. @logger fn demo(xx, xx) { }
type DecoratorExpr struct {
	Token     token.Token // '@'
	Decorator Expression  //Decorator function
	Decorated Expression  //Decorated function or another Decorator
}

func (dc *DecoratorExpr) Pos() token.Position {
	return dc.Token.Pos
}

func (dc *DecoratorExpr) End() token.Position {
	return dc.Decorated.End()
}

func (dc *DecoratorExpr) expressionNode()      {}
func (dc *DecoratorExpr) TokenLiteral() string { return dc.Token.Literal }
func (dc *DecoratorExpr) String() string {
	var out bytes.Buffer

	out.WriteString("@")
	out.WriteString(dc.Decorator.String())
	out.WriteString(" ")
	out.WriteString(dc.Decorated.String())

	return out.String()
}

type CmdExpression struct {
	Token token.Token
	Value string
}

func (c *CmdExpression) Pos() token.Position {
	return c.Token.Pos
}

func (c *CmdExpression) End() token.Position {
	length := utf8.RuneCountInString(c.Value)
	return token.Position{Filename: c.Token.Pos.Filename, Line: c.Token.Pos.Line, Col: c.Token.Pos.Col + length}
}

func (c *CmdExpression) expressionNode()      {}
func (c *CmdExpression) TokenLiteral() string { return c.Token.Literal }
func (c *CmdExpression) String() string       { return c.Value }
