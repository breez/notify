# notify
Notifications delivery tool that can be used both as a library and as an http service.
The package notify contains the generic structure and the ability to create a Notifier that uses underline Services to convert and deliver the notifications.

# Sample Usage
To create a Notifier one need to initialize as follows:
```
notifier := notify.NewNotifier(config, map[string]Service{"serviceType1": myServic1, "serviceType2": myServic2})
```

To send notification using FCM you can use the built-in FCM service:

```
fcm := services.NewFCM(createMessageFactory(), fcmClient)
notifier := notify.NewNotifier(c, map[string]notify.Service{
  "fcm": fcm,
}), nil
```

Then you can start sending a notification based on a specific template:

```
notification := Notification{
  Template: "template1",
  Type:     "fcm",
  Token:    "1234",
}
notifier.Notify(context.Background(), &notification)
```

You can also run it as an http service to allow for example webhooks as triggers for notifications:

```
http.Run(notifier, httpConfig)
```

  

