# GoFast

A framework to boost up your rest api app built with GoLang

```
app := gorun.New(&gorun.Config{})
app.Use(users.Setup(&users.Config{}))
app.Use(mailer.Setup())
app.Run()
```