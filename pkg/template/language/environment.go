// TemplateEngine environment
package language

import (
	"errors"
	"path"
)

// Environment. Template Parsing Environment
type Environment struct {
	engine  *TemplateEngine
	stack   *environmentStack
	options *Options
}

// newEnvironment creates a new environment object
func newEnvironment(engine *TemplateEngine, options *Options) *Environment {
	env := &Environment{
		engine:  engine,
		stack:   newStack(10),
		options: options,
	}

	return env
}

// currentDir returns the current directory of the file being parsed
func (env *Environment) currentDir() string {
	if env.stack.length() == 0 {
		return "."
	}

	currentFrame, _ := env.stack.Peek()
	return path.Dir(currentFrame.filePath)
}

// environment stack
type environmentStack struct {
	data []*environmentStackFrame
}

// environment stack frame
type environmentStackFrame struct {
	variables      map[string]interface{}
	fileRepository FileRepository
	filePath       string
}

// ErrEmptyStack raised when the stack is empty on pop or peek
var ErrEmptyStack = errors.New("stack.go : stack is empty")

func newStack(number uint) *environmentStack {
	return &environmentStack{data: make([]*environmentStackFrame, 0, number)}
}

// length return the number of items in stack
func (s *environmentStack) length() int {
	return len(s.data)
}

//Push pushes a frame into stack
func (s *environmentStack) Push(value *environmentStackFrame) {
	s.data = append(s.data, value)
}

//pop the top item out, if stack is empty, will return ErrEmptyStack decleared above
func (s *environmentStack) Pop() (*environmentStackFrame, error) {
	if s.length() > 0 {
		rect := s.data[s.length()-1]
		s.data = s.data[:s.length()-1]
		return rect, nil
	}
	return nil, ErrEmptyStack
}

//peek the top item. Notice, this is like a pointer:
//tmp, _ := s.Peek(); tmp = 123;
//s.Pop() should return 123, nil.
func (s *environmentStack) Peek() (*environmentStackFrame, error) {
	if s.length() > 0 {
		return s.data[s.length()-1], nil
	}
	return nil, ErrEmptyStack
}
