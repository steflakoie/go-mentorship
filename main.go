package main

import (
	"go-mentorship/module1/errchain"
)

// type CustomError struct {
// 	Code    int
// 	Message string
// }

// func (e CustomError) Error() string {
// 	return fmt.Sprintf("error: '%s' with a code: '%d'", e.Message, e.Code)
// }

// var ErrNotFound = errors.New("not found")
// var ErrNotFoundNumber = errors.New(strconv.Itoa(404))

// func main() {
// 	cerr := CustomError{
// 		Code:    404,
// 		Message: "Not found",
// 	}

// 	wrapedCerr := fmt.Errorf("wrapped err: %w", cerr)
// 	wrapedCerrTwice := fmt.Errorf("wrapped err twice: %w", wrapedCerr)

// 	fmt.Println(Is(wrapedCerr, cerr))
// 	fmt.Println(Is(wrapedCerrTwice, cerr))
// 	fmt.Println(errors.Is(wrapedCerrTwice, cerr))
// 	fmt.Println(errors.Is(wrapedCerrTwice, ErrNotFound))
// 	fmt.Println(Is(wrapedCerrTwice, ErrNotFound))

// }

// func Is(err, target error) bool {
// 	if err == nil {
// 		return false
// 	}

// 	if err == target {
// 		return true
// 	}

// 	value := reflect.ValueOf(err)
// 	unwrapMethod := value.MethodByName("Unwrap")
// 	isUnwrapMethodValid := unwrapMethod.IsValid()
// 	if isUnwrapMethodValid {
// 		// Получаю unwrap ошибку
// 		res := unwrapMethod.Call(nil)[0]
// 		if res.Kind() == reflect.Pointer || res.Kind() == reflect.Interface {
// 			if res.IsNil() {
// 				return false
// 			}
// 		}
// 		err = res.Interface().(error)

// 		return Is(err, target)
// 	}

// 	return false
// }

func main() {
	errchain.Errchain()
}
