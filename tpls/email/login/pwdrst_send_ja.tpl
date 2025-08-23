[{{HTML (T .Loc "site")}}] パスワードリセット

<div>
	<p>{{.User.Name}}様</p>
	<p>{{T .Loc "site"}}をご利用いただき、ありがとうございます。</p>
	<p>{{TIME .Now}}に&lt;{{.User.Email}}&gt;のログインパスワードのリセット要請を受付ました。</p>
	<p>パスワードをリセットしたい場合は、{{.Expires}}分以内に以下のリンクをタップしてください。</p>
	<p><a href="{{.ResetURL}}">{{.ResetURL}}</a></p>
	<br>
	<p>The {{T .Loc "site"}} Team.</p>
</div>
