package validator

import "fmt"

var ValidGrades = map[int]bool{
	5: true, 10: true, 15: true, 20: true,
	25: true, 30: true, 35: true, 40: true,
}

var ValidGradeSlice = []int{5, 10, 15, 20, 25, 30, 35, 40}

func ValidateFabricGrade(grade int) error {
	if !ValidGrades[grade] {
		return fmt.Errorf(
			"invalid fabric grade %d: allowed values are 5, 10, 15, 20, 25, 30, 35, 40",
			grade,
		)
	}
	return nil
}
