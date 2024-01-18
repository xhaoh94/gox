package types

type (
	AOIKey interface {
		Number | string
	}
	IAOIManager[T AOIKey] interface {
		Enter(T, float32, float32)
		Leave(T)
		Update(T, float32, float32)
		Find(T) IAOIResult[T]
	}
	IAOIResult[T AOIKey] interface {
		Owner() T
		Has(T) bool
		//不包括Owner
		IDList() []T
		//不包括Owner
		IDMap() map[T]bool
		Range(func(T))
		Reset()
		//(不包括Owner)Compare 比较  Complement补集 Minus差集 Intersect交集
		Compare(IAOIResult[T]) (Complement []T, Minus []T, Intersect []T)
	}
)
