.text
    func main() -> void {
        push byte 65
        syscall write_byte
        push byte 10
        syscall write_byte
    }
