package main

// なんか思ってたのと違う
// squareHandlerからmiddlewareを登録するわけだが、middlewareは常に実行されるので、squareHandlerの内容は常に実行されることになる。ハンドラの分岐ができなくないか。
// squareHandlerと/squareの対応付けは、ない。/square以外の場合は弾いているだけ
