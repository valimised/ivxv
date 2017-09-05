package errors

import (
	"fmt"
	"reflect"
)

func ExampleWalk() {
	err := WalkTopError{
		Err: WalkIntermediateError{
			Err: WalkRootError{}}}
	_ = Walk(err, func(err error) error {
		fmt.Printf("called f(%s)\n", reflect.TypeOf(err).Name())
		return nil
	})
	// Output:
	// called f(WalkTopError)
	// called f(WalkIntermediateError)
	// called f(WalkRootError)
}

func ExampleCausedBy() {
	var err error = CausedByTopError{
		Err: CausedByIntermediateError{
			Err: CausedByRootError{}}}
	fmt.Println("caused by top:", CausedBy(err, new(CausedByTopError)) != nil)
	fmt.Println("caused by intermediate:", CausedBy(err, new(CausedByIntermediateError)) != nil)
	fmt.Println("caused by root:", CausedBy(err, new(CausedByRootError)) != nil)
	fmt.Println("caused by other:", CausedBy(err, OtherError{}) != nil)
	// Output:
	// caused by top: true
	// caused by intermediate: true
	// caused by root: true
	// caused by other: false
}
