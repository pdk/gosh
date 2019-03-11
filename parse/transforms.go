package parse

import "github.com/pdk/gosh/token"

// transformFuncApply checks if the node is a function/method invocation, and
// rewrites as either FUNCAPPLY/f-apply or METHAPPLY/m-apply.
func (n *Node) transformFuncApply() *Node {

	// must be "blah ( ..." not just "( ...". And of course, if not "(" at all,
	// then we don't apply this transform.
	if n.Token() != token.LPAREN || !n.IsLefty() {
		return n
	}

	first := n.firstChild()
	if first == nil {
		return n
	}

	// transform method invocation
	// ("(" (. o m) ...) ==> (m-apply o m ...)
	if first.Token() == token.PERIOD && len(first.children) == 2 {

		n.lexeme = n.lexeme.Rewrite(token.METHAPPLY, "m-apply")
		n = n.raiseFirstChildren()

		return n
	}

	// transform plain function invocation
	// ("(" f ...) ==> (f-apply f ...)
	n.lexeme = n.lexeme.Rewrite(token.FUNCAPPLY, "f-apply")

	return n
}

func (n *Node) raiseSingleTuples() *Node {

	if n.Token() != token.LPAREN || len(n.children) != 1 {
		return n
	}

	return n.children[0].raiseSingleTuples()
}

// transformSemiTreeToList identifies the normal tree structure produced by
// parsing semicolons. It's really just a series of expressions, so convert it
// to a list. And also convert the token to STMS/stmts.
func (n *Node) transformSemiTreeToList() *Node {

	if n.Token() != token.SEMI {
		return n
	}

	// rename ";" to "stmts", cuz it's just a chain of statements, actually.
	// (; ...) ==> (stmts ...)
	n.lexeme = n.lexeme.Rewrite(token.STMTS, "stmts")

	first := n.firstChild()

	// unnest statements
	// (stmts (; (; (:= a 1) (:= b 2)) (:= c 3)))
	// ==> (stmts (:= a 1) (:= b 2) (:= c 3))
	if first != nil && (first.Token() == token.SEMI || first.Token() == token.STMTS) {
		n = n.raiseFirstChildren()
	}

	if len(n.children) == 1 {
		return n.children[0]
	}

	return n
}

// func (n *Node) transformCommaTreeToList() *Node {

// 	if n.Token() == token.COMMA && n.firstChild() != nil && n.firstChild().Token() == token.COMMA {
// 		return n.raiseFirstChildren()
// 	}

// 	return n
// }

// applyTransforms does some post-part, pre-eval transformations to "clean
// things up a bit".
func (n *Node) applyTransforms() *Node {

	n = n.transformFuncApply()
	n = n.raiseSingleTuples()

	for i, c := range n.children {
		n.children[i] = c.applyTransforms()
	}

	n = n.transformSemiTreeToList()
	// n = n.transformCommaTreeToList()

	return n
}

// raiseFirstChildren raises the children nodes of the first child to be
// children of this node.
// (x (y a b c ...) d e f ...) ==> (x a b c ... d e f ...)
func (n *Node) raiseFirstChildren() *Node {

	newChildren := n.children[0].children
	newChildren = append(newChildren, n.children[1:]...)
	n.children = newChildren

	return n
}
