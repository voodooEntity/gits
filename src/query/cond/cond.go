package cond

// Condition is an interface implemented by specific condition types.
type Condition interface {
	IsCondition() // Marker method
	SetNegated(negated bool) Condition
	IsNegated() bool
}

// LogicalOperator defines the type of logical grouping.
type LogicalOperator string

const (
	OpAnd LogicalOperator = "AND"
	OpOr  LogicalOperator = "OR"
)

// ConditionGroup represents a logical grouping of conditions (AND or OR).
type ConditionGroup struct {
	Type     LogicalOperator // "AND" or "OR"
	Operands []Condition     // A list of conditions (can be MatchCondition or other ConditionGroup)
	Negated  bool            // If true, the result of this group is negated
}

// IsCondition is a marker method for the Condition interface.
func (cg *ConditionGroup) IsCondition() {}

// SetNegated sets the negated status of the condition group and returns the modified group.
func (cg *ConditionGroup) SetNegated(negated bool) Condition {
	cg.Negated = negated
	return cg
}

// IsNegated returns the negated status of the condition group.
func (cg *ConditionGroup) IsNegated() bool {
	return cg.Negated
}

// MatchCondition represents a single comparison.
type MatchCondition struct {
	Field    string // e.g., "ID", "Value", "Context", "Properties.xyz"
	Operator string // e.g., "==", "!=", "prefix", "contain", ">", "<="
	Value    string // The value to compare against
	Negated  bool   // If true, the result of this match is negated
}

// IsCondition is a marker method for the Condition interface.
func (mc *MatchCondition) IsCondition() {}

// SetNegated sets the negated status of the match condition and returns the modified condition.
func (mc *MatchCondition) SetNegated(negated bool) Condition {
	mc.Negated = negated
	return mc
}

// IsNegated returns the negated status of the match condition.
func (mc *MatchCondition) IsNegated() bool {
	return mc.Negated
}

// Helper functions to construct conditions

// Match creates a new MatchCondition.
func Match(field, operator, value string) *MatchCondition {
	return &MatchCondition{
		Field:    field,
		Operator: operator,
		Value:    value,
		Negated:  false,
	}
}

// And creates a new ConditionGroup with AND logic.
func And(operands ...Condition) *ConditionGroup {
	return &ConditionGroup{
		Type:     OpAnd,
		Operands: operands,
		Negated:  false,
	}
}

// Or creates a new ConditionGroup with OR logic.
func Or(operands ...Condition) *ConditionGroup {
	return &ConditionGroup{
		Type:     OpOr,
		Operands: operands,
		Negated:  false,
	}
}
