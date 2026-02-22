package src

type Resurrect struct {
	Kitty Kitty
}

func InitResurrect(kitty Kitty) *Resurrect {
	return &Resurrect{
		Kitty: kitty,
	}
}

func (r *Resurrect) AddEventListener() {

}

func (r *Resurrect) SaveSession() {
	// var outputFile string = "~/.config/kitty/kitty-session.kitty"
}

func (r *Resurrect) Ressurect() {

}
