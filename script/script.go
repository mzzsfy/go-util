package script

import (
	"fmt"
)

// ========== Convenience Functions ==========

// Eval executes a script and returns the result.
// This is the simplest way to execute a script, suitable for one-time execution.
//
// Example:
//
//	result, err := script.Eval("10 + 20")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.Int()) // Output: 30
func Eval(source string) (Value, error) {
	parser := NewParser()
	defer returnParserToPool(parser)

	engine := NewEngine()
	ctx := NewContext()

	script, err := parser.Compile(source)
	if err != nil {
		return Value{}, err
	}

	return engine.Run(ctx, script)
}

// EvalWithBindings executes a script with variable bindings.
// Suitable for scenarios requiring external data.
//
// Example:
//
//	bindings := map[string]interface{}{
//	    "name": "Alice",
//	    "age": 30,
//	}
//	result, err := script.EvalWithBindings(`
//	    n := getBindValue("name")
//	    a := getBindValue("age")
//	    "Hello, " + n + "! Age: " + string(a)
//	`, bindings)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.String()) // Output: Hello, Alice! Age: 30
func EvalWithBindings(source string, bindings map[string]interface{}) (Value, error) {
	parser := NewParser()
	defer returnParserToPool(parser)

	engine := NewEngine()
	ctx := NewContext()

	// Bind all variables
	for name, value := range bindings {
		ctx.BindValue(name, value)
	}

	script, err := parser.Compile(source)
	if err != nil {
		return Value{}, err
	}

	return engine.Run(ctx, script)
}

// MustEval executes a script and panics if an error occurs.
// Suitable for executing scripts during initialization that are guaranteed not to fail.
// Similar to template.Must in the Go standard library.
//
// Example:
//
//	// During initialization
//	config := script.MustEval(`{"timeout": 30, "retries": 3}`)
//	fmt.Println(config.Map().Pairs["timeout"].Int()) // Output: 30
func MustEval(source string) Value {
	result, err := Eval(source)
	if err != nil {
		panic(fmt.Sprintf("script.Eval failed: %v", err))
	}
	return result
}

// MustEvalWithBindings executes a script with bindings and panics if an error occurs.
// Suitable for executing scripts during initialization that are guaranteed not to fail.
//
// Example:
//
//	bindings := map[string]interface{}{"x": 10, "y": 20}
//	result := script.MustEvalWithBindings("x + y", bindings)
//	fmt.Println(result.Int()) // Output: 30
func MustEvalWithBindings(source string, bindings map[string]interface{}) Value {
	result, err := EvalWithBindings(source, bindings)
	if err != nil {
		panic(fmt.Sprintf("script.EvalWithBindings failed: %v", err))
	}
	return result
}
