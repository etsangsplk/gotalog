package gotalog

import "fmt"

type memClauseStore map[string]clause

func (mem memClauseStore) set(key string, c clause) error {
	mem[key] = c
	return nil
}

func (mem memClauseStore) delete(key string) error {
	delete(mem, key)
	return nil
}

func (mem memClauseStore) size() (int, error) {
	return len(mem), nil
}

func (mem memClauseStore) iterator() (chan clause, error) {
	clauses := make(chan clause, 1)
	go func() {
		for _, c := range mem {
			clauses <- c
		}
		close(clauses)
	}()
	return clauses, nil

}

type memDatabase map[string]*predicate

// TODO: we need to somehow intern predicates on the basis of string/int identification,
// so that multiple clauses referring to the same predicate can reach the same
// db of clauses.
func (db memDatabase) newPredicate(n string, a int) *predicate {

	p := &predicate{
		Name:      n,
		Arity:     a,
		db:        memClauseStore{},
		primitive: nil,
	}
	if existing, ok := db[p.getID()]; ok {
		return existing
	}
	db[p.getID()] = p
	return p
}

func (db memDatabase) insert(pred *predicate) {
	db[pred.getID()] = pred
}

func (db memDatabase) remove(pred predicate) predicate {
	delete(db, pred.getID())
	return pred
}

// assertions should only be made for clauses' whose
// predicates originate within the same database.
func (db memDatabase) assert(c clause) error {
	if !isSafe(c) {
		return fmt.Errorf("cannot assert unsafe clauses")
	}

	pred := c.head.pred
	// Ignore assertions on primitive predicates
	if pred.primitive != nil {
		return fmt.Errorf("cannot assert on primitive predicates")
	}
	return pred.db.set(c.getID(), c)
}

func (db memDatabase) retract(c clause) error {
	pred := c.head.pred
	err := pred.db.delete(c.getID())
	if err != nil {
		// This leads to garbage in the predicate's database.
		return err
	}

	// If a predicate has no clauses associated with it, remove it from the db.
	size, err := pred.db.size()
	if err != nil {
		// Likewise, we end up with garbage if this happens.
		return err
	}

	if size == 0 {
		db.remove(*pred)
	}
	return nil
}