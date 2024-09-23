package main

import (
	"fmt"
	"time"
)

// Constants for HSA calculations
const (
	CatchUpAge          = 55
	CatchUpContribution = 1000
)

// HSA-related limits from https://www.kiplinger.com/taxes/hsa-contribution-limit-2024#:~:text=Contribution%20Limits%20for%20Health%20Savings%20Accounts%20Under%20HDHPs
var (
	HDHPMinimumDeductible = map[string]int{
		"Self-only": 1600,
		"Family":    3200,
	}
	HSAContributionLimit = map[string]int{
		"Self-only": 4150,
		"Family":    8300,
	}
)

// Employee represents an individual employee's data
type Employee struct {
	Name               string `json:"Name"`
	PlanType           string `json:"Plan Type"`
	Deductible         int    `json:"Deductible"`
	DateOfBirth        string `json:"Date of birth"`
	HSAEligible        bool
	HSAMaxContribution int
}

// calculateAge calculates the age of an employee based on their date of birth
func calculateAge(dateOfBirth string) (int, error) {
	dob, err := time.Parse("2006-01-02", dateOfBirth)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	age := now.Year() - dob.Year()

	// Adjust age if birthday hasn't occurred this year
	if now.YearDay() < dob.YearDay() {
		age--
	}

	return age, nil
}

// IsHSAEligible determines if an employee is eligible for an HSA based on their plan type and deductible
func IsHSAEligible(planType string, deductible int) bool {
	return deductible >= HDHPMinimumDeductible[planType]
}

// CalculateHSAMaxContribution calculates the maximum HSA contribution for an employee
func CalculateHSAMaxContribution(planType, dateOfBirth string) (int, error) {
	maxContribution, exists := HSAContributionLimit[planType]
	if !exists {
		return 0, fmt.Errorf("unknown plan type: %s", planType)
	}

	age, err := calculateAge(dateOfBirth)
	if err != nil {
		return 0, err
	}

	// Add catch-up contribution for employees 55 and older
	if age >= CatchUpAge {
		maxContribution += CatchUpContribution
	}

	return maxContribution, nil
}

// ProcessEmployeeData calculates HSA eligibility and maximum contributions for each employee
func ProcessEmployeeData(employees []Employee) []Employee {
	for i, emp := range employees {
		employees[i].HSAEligible = IsHSAEligible(emp.PlanType, emp.Deductible)
		if employees[i].HSAEligible {
			maxContribution, err := CalculateHSAMaxContribution(emp.PlanType, emp.DateOfBirth)
			if err != nil {
				fmt.Printf("Error calculating HSA max contribution for %s: %v\n", emp.Name, err)
				employees[i].HSAMaxContribution = 0
			} else {
				employees[i].HSAMaxContribution = maxContribution
			}
		} else {
			employees[i].HSAMaxContribution = 0
		}

		// Debug output for each employee
		fmt.Printf("Employee: %s, Plan Type: %s, Deductible: %d, DOB: %s, HSA Eligible: %v, Max Contribution: %d\n",
			emp.Name, emp.PlanType, emp.Deductible, emp.DateOfBirth, employees[i].HSAEligible, employees[i].HSAMaxContribution)
	}
	return employees
}