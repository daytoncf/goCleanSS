package css

import (
	"strings"

	lib "github.com/daytoncf/goCleanSS/pkg/lib"
)

type TokenType int

const (
	COMMENT TokenType = iota
	RULESET
	ERR
)

type AtRuleType int

const (
	CHARSET AtRuleType = iota
	COUNTERSTYLE
	FONTFACE
	IMPORT
	KEYFRAMES
	MEDIA
	PAGE
	SUPPORTS
	ATERROR
)

type Declaration struct {
	Property string
	Value    string
}

type Token struct {
	TokenType    TokenType
	Selector     string
	Declarations []Declaration
}

type AtRule struct {
	AtRuleType AtRuleType
	Selector   string
	Tokens     []Token
}

// Factory function for AtRule
func NewAtRule(t AtRuleType, selector string, tokens []Token) AtRule {
	return AtRule{t, selector, tokens}
}

// Factory function for CSSToken
func NewToken(t TokenType, selector string, declarations []Declaration) Token {
	return Token{t, selector, declarations}
}

// Factory function for CSSDeclaration
func NewDeclaration(property, value string) Declaration {
	return Declaration{property, value}
}

func ParseDeclarationBlock(declarationBlock string) []Declaration {
	// Initialize empty array
	declarations := make([]Declaration, 0)
	// Trim trailing and leading whitespace.
	minDeclarationBlock := strings.TrimSpace(declarationBlock)

	// Initialize temp variables to store parsed values
	var tempProperty, tempValue string
	var charQueue []rune
	for _, char := range minDeclarationBlock {
		switch char {
		case ':':
			tempProperty = strings.TrimSpace(lib.PopRuneArrToString(&charQueue))
		case ';':
			tempValue = strings.TrimSpace(lib.PopRuneArrToString(&charQueue))
			declarations = append(declarations, NewDeclaration(tempProperty, tempValue))
		default:
			charQueue = append(charQueue, char)
		}
	}
	return declarations
}

// Function that parses an entire @rule block, which will have its own set of tokens
func ParseAtRuleBlock(atRuleBlock string) []Token {
	// Initialize tokens to be returned
	tokens := make([]Token, 0)

	var charQueue lib.Queue
	var readingComment bool = false

	var selector string
	var decBlock string
	for i, v := range atRuleBlock {
		switch v {
		case '/':
			if readingComment && peekForCommentEnd(atRuleBlock, i) {
				comment := strings.TrimSpace(charQueue.PopQueueToString())
				// Create Token for comment, using its contents for the selector exluding asterisk, [1:len(s)-1]
				tokens = append(tokens, NewToken(COMMENT, comment, []Declaration{}))
			} else {
				// Check to see if there is an asterisk following this /
				readingComment = peekForCommentStart(atRuleBlock, i)
			}
		case '{':
			selector = strings.TrimSpace(charQueue.PopQueueToString())
		case '}':
			decBlock = strings.TrimSpace(charQueue.PopQueueToString())
			// Create new token after declaration block finishes :)
			tokens = append(tokens, NewToken(RULESET, selector, ParseDeclarationBlock(decBlock)))
		default:
			charQueue.Push(v)
		}
	}

	return tokens
}

// Function that takes full selector, such as `@media screen and (max-width: 600px)` and returns the atRule type.
// In the above example, the function will return `MEDIA` for @media.
func getAtRuleType(selector string) AtRuleType {
	atRuleName := selector[:strings.Index(selector, " ")]

	switch atRuleName {
	case "@charset":
		return CHARSET
	case "@counter-style":
		return COUNTERSTYLE
	case "@font-face":
		return FONTFACE
	case "@import":
		return IMPORT
	case "@keyframes":
		return KEYFRAMES
	case "@media":
		return MEDIA
	case "@page":
		return PAGE
	case "@supports":
		return SUPPORTS
	}
	return ATERROR // ahhhhh it didnt work!! its a terror!!! (this is a joke bad about my bad naming)
}

// Removes all whitespace characters within a given string
func Tokenizer(path string) ([]Token, []AtRule) {
	var tokens []Token
	var atRules []AtRule
	// Convert file into string to make it easily iterable
	fileString := lib.FileToString(path)

	var charQueue lib.Queue
	var charStack lib.Stack
	var readingComment, readingAtRule bool = false, false
	var comment, selector, atRuleSelector, decBlock, atRuleBlock string
	for i, v := range fileString {
		switch v {
		case '/':
			if readingComment && peekForCommentEnd(fileString, i) {
				// Pop comments contents into `comment`
				comment = strings.TrimSpace(charQueue.PopQueueToString())
				// Create Token for comment, using its contents for the selector exluding asterisk, [1:len(s)-1]
				tokens = append(tokens, NewToken(COMMENT, comment, []Declaration{}))
				readingComment = false
			} else {
				// Check to see if there is an asterisk following this /
				readingComment = peekForCommentStart(fileString, i)
			}
		case '{':

			if charStack.IsEmpty() {
				// Pop selector name into `selector`
				selector = strings.TrimSpace(charQueue.PopQueueToString())

				if strings.HasPrefix(selector, "@") {
					// If the selector is an @rule, hold onto this. We will use it later
					atRuleSelector = selector
				}
			} else {
				charQueue.Push(v) // Include the braces in this block, will use for parsing later
			}
			// Push '{' onto charStack to keep track of nested / @rule blocks
			charStack.Push(v)
		case '}':
			// Pop '{' off top of stack
			charStack.Pop()

			if charStack.IsEmpty() && readingAtRule {
				readingAtRule = false
				atRuleBlock = strings.TrimSpace(charQueue.PopQueueToString())
				atRules = append(atRules, NewAtRule(getAtRuleType(atRuleSelector), atRuleSelector, ParseAtRuleBlock(atRuleBlock)))
			} else if charStack.IsEmpty() {
				// Pop contents of the declaration block into `decBlock`
				decBlock = strings.TrimSpace(charQueue.PopQueueToString())

				// Create new token after declaration block finishes :)
				tokens = append(tokens, NewToken(RULESET, selector, ParseDeclarationBlock(decBlock)))
			} else {
				readingAtRule = true
				charQueue.Push(v) // Include the braces in this block, will use for parsing later
			}

		default:
			charQueue.Push(v)
		}
	}
	return tokens, atRules
}

// Function that is called after a '/' rune is encountered.
// Check to see if the following rune in string is an asterisk. Returns true if so.
func peekForCommentStart(fileString string, currPos int) bool {
	return fileString[currPos+1] == '*'
}

// Function that is called after a '/' rune is encountered.
// Check to see if the preceding rune in string is an asterisk. Returns true if so.
func peekForCommentEnd(fileString string, currPos int) bool {
	return fileString[currPos-1] == '*'
}
