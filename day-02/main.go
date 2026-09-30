package main

import "fmt"

func calculate(a, b float64, op byte) (float64, error) {
	switch op {
	case '+':
		return a + b, nil
	case '-':
		return a - b, nil
	case '*':
		return a * b, nil
	case '/':
		if b == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return a / b, nil
	case '%':
		if b == 0 {
			return 0, fmt.Errorf("modulo by zero")
		}
		return float64(int(a) % int(b)), nil
	default:
		return 0, fmt.Errorf("unknown operator: %c", op)
	}
}

func main() {
	cases := []struct {
		a, b float64
		op   byte
	}{
		{12, 4, '+'},
		{12, 4, '-'},
		{12, 4, '*'},
		{12, 4, '/'},
		{12, 4, '%'},
		{12, 0, '/'},
	}

	for _, c := range cases {
		result, err := calculate(c.a, c.b, c.op)
		if err != nil {
			fmt.Printf("%.2f %c %.2f = error: %v\n", c.a, c.op, c.b, err)
		} else {
			fmt.Printf("%.2f %c %.2f = %.2f\n", c.a, c.op, c.b, result)
		}
	}
}
