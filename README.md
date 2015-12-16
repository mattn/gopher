# Gopher

![](http://i.imgur.com/9Xf0BTM.gif)

Love Gopher

This is a desktop mascot application running on your desktop on windows.

## Feature

At the first, you need to run `gopher.exe`.

### Walking

![](http://i.imgur.com/BgiIAj9.gif)

So Sweet!

### SL

If you did mis-type `ls` as `sl`.

![](http://i.imgur.com/xt550tv.gif)

So Fast!

### Notification

This repository bundled `gopherc.exe` that is client application to operate Gopher.

```
gopherc -m Hello
```

![](http://i.imgur.com/DTIBM9W.gif)

Hello Gopher!

### Jumping

```
gopherc -j
```

![](http://i.imgur.com/OKqbF7n.gif)

Looking Good!

### Vim plugin

Use `misc/vim` if you are vimmer.

```
:HeyGopher おなかすいた
```

![](http://i.imgur.com/K9h25F5.png)

So don't worry even if you are remaining alone at your office and you feel lonely.

### RSS Notification

Run `gopherfeed.exe` to aggregate RSS feed.

![](http://i.imgur.com/UEdmHYI.png)

It's Nice!

### Websocket Chat

![](http://i.imgur.com/PMVBSJ2.png)

It seems that the members of the chat room are talking on your windows desktop.

## Requirements

No, this is fully statically executable file. And this's not CGO.
But unfortunately, this only works on windows.

## Installation

```
cd cmd\gopher
mingw32-make
```
And copy `gopher.exe` into the path which is contained in %PATH% environment variables.

## License

MIT

The image files of gopher is created by Renee French.

## Author

Yasuhiro Matsumoto (a.k.a mattn)
