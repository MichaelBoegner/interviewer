package chatgpt

import "fmt"

func GetFirstQuestion(username string) (string, error) {
	return fmt.Sprintf("%v, for your first question, what is the airspeed of an unladdened swallow?", username), nil
}
