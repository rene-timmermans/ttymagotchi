package app

// The same expression is used on Status, during care and on result screens.
type petExpression struct{ eyes, mouth, brow string }

func expressionFor(happiness int, alive bool) petExpression {
	if !alive {
		return petExpression{"x", "-", " "}
	}
	switch {
	case happiness >= 80:
		return petExpression{"^", "u", "^"}
	case happiness >= 60:
		return petExpression{"o", "w", "^"}
	case happiness >= 40:
		return petExpression{"o", "-", " "}
	case happiness >= 20:
		return petExpression{".", "~", "/"}
	case happiness > 0:
		return petExpression{",", "n", "/"}
	default:
		return petExpression{";", "n", "/"}
	}
}

func (e petExpression) rightBrow() string {
	if e.brow == "/" {
		return "\\"
	}
	return e.brow
}
