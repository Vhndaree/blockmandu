package main

func main() {
	bc := NewBlockChain()
	defer bc.db.Close()

	NewCli(bc).Run()
}
