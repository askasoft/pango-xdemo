[{{HTML (T .Loc "site")}}] 二段階認証のお知らせ

<div>
	<p>{{T .Loc "site"}}をご利用いただき、ありがとうございます。</p>
	<p>{{TIME .Now}}にログインリクエストが受信されました。</p>
	<p>{{.Expires}}分以内に下記の認証コードをログイン画面に入力して、二段階認証を続行してください。</p>
	<p>認証コード: {{.Passcode}}</p>
	<br>
	<p>The {{T .Loc "site"}} Team.</p>
</div>
