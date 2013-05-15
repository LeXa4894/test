package hello

import (
    "appengine"
    "appengine/datastore"
    "appengine/user"
    "html/template"
    "net/http"
	"strconv"
    "fmt"
	"time"
)

type User struct{
	Name string
}

type UserC struct{
	Name string
	Cost float64
}

type Pay struct{
	Cost	float64
	Date    time.Time
	//Comment string
	Payer	string
}

type Trans struct{
	Cost	float64
	Date    time.Time
	//Comment string
	From	string
	To		string
}



func init() {
    http.HandleFunc("/", root)
    http.HandleFunc("/AddT",addT)
    http.HandleFunc("/AddTB",AddTBackend)
    http.HandleFunc("/AddP",addP)
    http.HandleFunc("/AddPB",AddPBackend)
    http.HandleFunc("/See",See)
    http.HandleFunc("/status",status)
    http.HandleFunc("/admin",admin)
    http.HandleFunc("/admin/Add",Add)
    http.HandleFunc("/admin/Clean",Clean)
}

const AdminTemplateHTML = `
<html>
  <body>
	<p> Add user </p>
    <form action="/admin/Add" method="post">
      <div><textarea name="username" rows="1" cols="60"></textarea></div>
      <div><input type="submit" value="Add user"></div>
    </form>
    <p> List users: </p>
    {{range .}}
    <p><b>{{.Name}}</b></p>
    {{end}}
    <p> Clean data </p>
    <form action="/admin/Clean" method="post">
      <div><input type="submit" value="Clean"></div>
    </form>
  </body>
</html>
`
var AdminTemplate = template.Must(template.New("book").Parse(AdminTemplateHTML))


func admin(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)
    u := user.Current(c)
    if !u.Admin {
        url, err := user.LoginURL(c, r.URL.String())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Location", url)
        w.WriteHeader(http.StatusFound)
        return
    }
    q := datastore.NewQuery("User").Limit(100)
    k := make([]User,0,100)
    _,err := q.GetAll(c,&k); 
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    if err := AdminTemplate.Execute(w, k); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func Clean(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
	q := datastore.NewQuery("User").Limit(100)
	x := make([]User,0,100)
	k , err:= q.GetAll(c,&x)
	if  err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	datastore.DeleteMulti(c,k)
	q2 := datastore.NewQuery("Pay").Limit(100)
	x2 := make([]Pay,0,100)
	k2 , err2:= q2.GetAll(c,&x2)
	if  err2 != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	datastore.DeleteMulti(c,k2)
	q3 := datastore.NewQuery("Trans").Limit(100)
	x3 := make([]Trans,0,100)
	k3 , err3:= q3.GetAll(c,&x3)
	if  err3 != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	datastore.DeleteMulti(c,k3)
    
    http.Redirect(w, r, "/admin", http.StatusFound)
}

func Add(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
	n := User{
		Name: r.FormValue("username"),
	}
    _, err := datastore.Put(c, datastore.NewIncompleteKey(c, "User", nil), &n)
    if  err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/admin", http.StatusFound)
}

const rootTemplateHTML = `
<html>
  <body>
	<form action="/AddT" method="get">
		<input type="submit" value="Add transfer">
	</form>
	<form action="/AddP" method="get">
		<input type="submit" value="Add pay">
	</form>
	<form action="/See" method="get">
		<input type="submit" value="See last operations">
	</form>
	<form action="/status" method="get">
		<input type="submit" value="See status">
	</form>
	<form action="/admin" method="get">
		<input type="submit" value="Go to admin">
	</form>
  </body>
</html>
`

func root(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    u := user.Current(c)
    if u == nil {
        url, err := user.LoginURL(c, r.URL.String())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Location", url)
        w.WriteHeader(http.StatusFound)
        return
    }
    
	fmt.Fprint(w,rootTemplateHTML)
}

const AddTTemplateHTML = `
<html>
  <body>
	<form action="/AddTB" method="post">
	<p>
	From:
	<select name="from" size="1">
	{{range .}}
		<option value="{{.Name}}">{{.Name}}</option>
	{{end}}
	</select>
	</p>
	<p>
	To:
	<select name="to" size="1">
	{{range .}}
		<option value="{{.Name}}">{{.Name}}</option>
	{{end}}
	</select>
	</p>
	<p>
	Some:
		<input type="text" name="cost">
		</p>
	<p>
	Date: <input type="date" name="date">
	</p>
	<input type="submit" value="Add transfer">
	</form>
  </body>
</html>
`
var AddTTemplate = template.Must(template.New("book").Parse(AddTTemplateHTML))

func addT(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)
	q := datastore.NewQuery("User").Limit(100)
    u := make([]User,0,100)
    _,err := q.GetAll(c,&u); 
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
	if err := AddTTemplate.Execute(w, u); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func AddTBackend(w http.ResponseWriter, r *http.Request){
	d,_:=strconv.ParseFloat(r.FormValue("cost"),64)
	t,_:=time.Parse("2006-01-02",r.FormValue("date"))
	c := appengine.NewContext(r)
	
	n := Trans{
		
		Cost: 	d,
		From: 	r.FormValue("from"),
		To: 	r.FormValue("to"),
		Date: 	t,
	}
    _, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Trans", nil), &n)
	if  err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	http.Redirect(w, r, "/AddT", http.StatusFound)
}
const AddPTemplateHTML = `
<html>
  <body>
	<form action="/AddPB" method="post">
	<p>
	Payer:
	<select name="payer" size="1">
	{{range .}}
		<option value="{{.Name}}">{{.Name}}</option>
	{{end}}
	</select>
	</p>
	<p>
	Some:
		<input type="text" name="cost">
		</p>
	<p>
	Date: <input type="date" name="date">
	</p>
	<input type="submit" value="Add pay">
	</form>
  </body>
</html>
`
var AddPTemplate = template.Must(template.New("book").Parse(AddPTemplateHTML))
func addP(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)
	q := datastore.NewQuery("User").Limit(100)
    k := make([]User,0,100)
    _,err := q.GetAll(c,&k); 
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
	if err := AddPTemplate.Execute(w, k); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func AddPBackend(w http.ResponseWriter, r *http.Request){
	d,_:=strconv.ParseFloat(r.FormValue("cost"),64)
	t,_:=time.Parse("2006-01-02",r.FormValue("date"))
	c := appengine.NewContext(r)
	
	n := Pay{
		Cost: 	d,
		Payer: 	r.FormValue("payer"),
		Date: 	t,
	}
    _, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Pay", nil), &n)
	if  err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	http.Redirect(w, r, "/AddP", http.StatusFound)
}

const SeePTemplateHTML = `
<html>
  <body>
	Pays:
	{{range .}}
	<p>
		{{.Date}}  -  {{.Cost}}р.  - {{.Payer}}
	</p>
	{{end}}

`
var SeePTemplate = template.Must(template.New("book").Parse(SeePTemplateHTML))

const SeeTTemplateHTML = `
	Transfers:
	{{range .}}
	<p>
		{{.Date}}  -  {{.Cost}}р.  -  {{.From}}  ->  {{.To}}
	</p>
	{{end}}
	<form action="/" method="post">
	<input type="submit" value="Back">
	</form>
  </body>
</html>
`
var SeeTTemplate = template.Must(template.New("book").Parse(SeeTTemplateHTML))
func See(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Pay").Order("-Date").Limit(100)
    k := make([]Pay,0,100)
    _,err1 := q.GetAll(c,&k); 
    //fmt.Fprintf(w,len(k))
    if err1 != nil {
            http.Error(w, err1.Error(), http.StatusInternalServerError)
            return
        }
	if err := SeePTemplate.Execute(w, k); err != nil {
        http.Error(w, err1.Error(), http.StatusInternalServerError)
    }
    q2:= datastore.NewQuery("Trans").Limit(100)
    l := make([]Trans,0,100)
    _,err2 := q2.GetAll(c,&l); 
    //fmt.Fprintf(w,len(l))
    if err2 != nil {
            http.Error(w, err2.Error(), http.StatusInternalServerError)
            return
        }
    if err := SeeTTemplate.Execute(w, l); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

const StatusTemplateHTML = `
<html>
  <body>
	{{range .}}
	<p>
		{{.Name}} - {{.Cost}}
	</p>
	{{end}}
	<form action="/" method="post">
	<input type="submit" value="Back">
	</form>
  </body>
</html>
`
var StatusTemplate = template.Must(template.New("book").Parse(StatusTemplateHTML))
func status(w http.ResponseWriter, r *http.Request){
	c := appengine.NewContext(r)
	q := datastore.NewQuery("User").Limit(100)
    k1 := make([]User,0,100)
    _,err := q.GetAll(c,&k1); 
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    
     qa :=datastore.NewQuery("Pay");
     ca,_:=q.Count(c)
     k2 := make([]Pay,0,ca)
     _,err = qa.GetAll(c,&k2); 
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    cost:=0.0;
    for _, valuea := range k2{
		cost+=valuea.Cost
	}
	cost_1 := cost/float64(ca)
    //fmt.Fprintf(w,"All cost : %f\n",cost_1)
    k :=make([]UserC,0,len(k1))
    for _, value := range k1{
		t := UserC{value.Name,0}
		q1 := datastore.NewQuery("Pay").Filter("Payer=",value.Name)
		c1,_:=q1.Count(c)
		//fmt.Fprintf(w,"%s pays Count : %d\n",value.Name,c1)
		t1 := make([]Pay,0,c1)
		_,err1 := q1.GetAll(c,&t1); 
		if err1 != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		for _, value1 := range t1{
			t.Cost+=value1.Cost
			//fmt.Fprintf(w,"%s pay Cost :        %f\n",value.Name,value1.Cost)
		}
		q2 := datastore.NewQuery("Trans").Filter("From =",value.Name)
		c2,_:=q2.Count(c)
		//fmt.Fprintf(w,"%s from transfers Count : %d\n",value.Name,c2)
		t2 := make([]Trans,0,c2)
		_,err2 := q2.GetAll(c,&t2); 
		if err2 != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		for _, value2 := range t2{
			t.Cost+=value2.Cost
			//fmt.Fprintf(w,"%s from transfer Cost :        %f\n",value.Name,value2.Cost)
		}
		q3 := datastore.NewQuery("Trans").Filter("To =",value.Name)
		c3,_:=q3.Count(c)
		//fmt.Fprintf(w,"%s to transfers Count : %d\n",value.Name,c3)
		t3 := make([]Trans,0,c3)
		_,err3 := q3.GetAll(c,&t3); 
		if err3 != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		for _, value3 := range t3{
			t.Cost-=value3.Cost
			//fmt.Fprintf(w,"%s to transfer Cost :        %f\n",value.Name,value3.Cost)
		}
		t.Cost-=cost_1
		//fmt.Fprintf(w,"%s Cost :        %f\n",value.Name,t.Cost-cost_1)	
		k=append(k,t)
	}
	//TODO dodelat
	
	if err := StatusTemplate.Execute(w, k); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}



