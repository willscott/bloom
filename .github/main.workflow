workflow "New workflow" {
  resolves = ["cedrickring/golang-action@1.2.0"]
  on = "push"
}

action "cedrickring/golang-action@1.2.0" {
  uses = "cedrickring/golang-action@1.2.0"
}
