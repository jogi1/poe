  package poe

  import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"errors"
)

type Character struct {
	Name            string
	League          string
	ClassId         int
	AscendancyClass int
	Class           string
	Level           int
}

type Item struct {
	Ilvl int
	Verified bool
	Name string
	CosmeticMods []string
	X int
	Y int
	TypeLine string
	SocketedItems []Item
}

type CharacterItems struct {
	Items []Item
	Character Character
}

type Tab struct {
	N string
	I int
}

type Stash struct {
	Items []Item
	NumTabs int
	Tabs []Tab
}

type Poe struct {
  accountName, email, password string
  client http.Client
  loggedIn bool
  loginHash string
}

func (P *Poe) Login (AccountName string, Email string, Password string) error {
  P.accountName = AccountName
  P.email = Email
  P.password = Password
  P.loggedIn = false

	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

  // get login hash
	P.client = http.Client{Jar: jar, CheckRedirect: nil}
	resp, err := P.client.Get("https://www.pathofexile.com/login/")
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	hashRegex, _ := regexp.Compile("name=\"hash\" value=\"(.*)\" ")
	str := string(data[:])
	hash := hashRegex.FindStringSubmatch(str)
	if hash == nil {
		return errors.New("error retrieving login hash")
	}
  P.loginHash = hash[1]

  // logging in
	resp, err = P.client.PostForm("https://www.pathofexile.com/login/", url.Values{
		"login_email":    {P.email},
		"login_password": {P.password},
		"login":          {"Login"},
		"remember_me":    {"0"},
		"hash":           {P.loginHash},
	})

	if err != nil {
		return err
	}

	data, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	invalidLoginRegex, _ := regexp.Compile("Invalid Login")
	if invalidLoginRegex.MatchString(string(data)) {
		return errors.New("Invalid Login (email or password might be wrong)")
	}

  P.loggedIn = true
  return nil
}

func (P *Poe) GetCharacters() (characters []Character, err error) {
  if P.loggedIn == false {
    return characters, errors.New(fmt.Sprintf("not logged in %v", P))
  }
	resp, err := P.client.Get("https://www.pathofexile.com/character-window/get-characters")
	if err != nil {
		return characters, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return characters, err
	}

	characters = make([]Character, 0)
	json.Unmarshal(data, &characters)

	return characters, nil
}

func (P *Poe) GetCharacterItems(characterName string) (characterItems CharacterItems, err error) {
  if !P.loggedIn {
    return characterItems, errors.New("not logged in")
  }
	resp, err := P.client.PostForm("https://www.pathofexile.com/character-window/get-items", url.Values{
		"accountName": {P.accountName},
		"character":   {characterName},
	})
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	json.Unmarshal(data, &characterItems)
	return characterItems, nil
}

func (P *Poe) GetStash(leagueName string, tabIndex int, tabs int) (stash Stash, err error) {
  if !P.loggedIn {
    return stash, errors.New("not logged in")
  }
	resp, err := P.client.PostForm("https://www.pathofexile.com/character-window/get-stash-items", url.Values{
		"accountName": {P.accountName},
		"league":   {leagueName},
		"tabIndex":   {fmt.Sprintf("%v", tabIndex)},
		"tabs":   {fmt.Sprintf("%v", tabs)},
	})
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	json.Unmarshal(data, &stash)
	return stash, nil
}
