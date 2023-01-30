package aoibase

type (
	IAOIResult interface {
		IDList() []string
		IDMap() map[string]bool
		Range(func(string))
		Reset()
		//Compare 比较  Complement补集 Minus差集 Intersect交集
		Compare(IAOIResult) (Complement []string, Minus []string, Intersect []string)
	}
)
