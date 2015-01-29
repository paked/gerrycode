const LOGIN = "Login";
const REGISTER = "Register";

class LoginCtrl {
	constructor() {
		this.mode = REGISTER;
	}

	go() {
		var url = "/api/user/login";
		if (this.mode == REGISTER) {
			url = "/api/user/create";
		}

		$.ajax({
			url: url,
			type: "POST",
			data: {
				username: $("#username").val(),
				password: $("#password").val(),
				email: $("#email").val()
			}
		}).
		done(msg => {
			if (msg.status.error) {
				this.error = msg.message;
				return
			}

			window.token = msg.token;
			$(location).attr("href", "#/r/pineapples");
		});
	}

	changeMode() {
		if (this.mode == LOGIN) {
			this.mode = REGISTER;
		} else {
			this.mode = LOGIN;
		}
	}

	otherMode() {
		return this.mode == "Login" ? "Register" : "Login";
	}

}

export default LoginCtrl;