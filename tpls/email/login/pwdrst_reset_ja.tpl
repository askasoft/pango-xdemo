<s>[{{T .Loc "sitename"}}] セキュリティー通知</s>

<div>
	<p>{{.User.Name}}様</p>
	<p>{{T .Loc "sitename"}}をご利用いただき、ありがとうございます。</p>
	<p>{{TIME .Now}}に&lt;{{.User.Email}}&gt;のログインパスワードがリセットされました。</p>
	<p>ご自身による操作であれば、何もする必要がありません。</p>
	<p>該当操作に心当たりがない場合は、速やかににログインパスワードを変更してください。</p>
	<br>
	<p>The {{T .Loc "sitename"}} Team.</p>
</div>
