.structs 
    struct person {
        age: int32
    }
.text
    func newPerson(age: int32) -> person {
        store 0
        newstruct person 
        dup
        load 0
        stfield "age"
        ret
    }
    func main() -> void {
        push int32 40
        call newPerson
        fldget "age"
        ret
    }
