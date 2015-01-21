class HomeView extends Backbone.View {

  initialize () {
    this.template = $('script[name="home"]').html();
  }

  render () {
    this.$el.html(_.template(this.template));
    return this;
  }

}

class LoginView extends Backbone.View {
	initialize() {
		this.template = $('script[name="login"]').html();
		this.events = {
			"click #login_button": "login"
		}	
	}

	login () {
		console.log("HEY HEY EY")
		$.ajax({
			url: "/api/user/login",
			type: "POST",
			dataType:"json",
			data: {
				username: $("#username_field").val(),
				password: $("#password_field").val()
			},
			success: function(data) {
				console.log(data, "HIlo")
				window.token = data.value

				$(location).attr("href", "#/")
			},
			error: function(xhr, status, error) {
				console.log("We had an error: ", xhr, status, error)
			}
		})
	}

	render () {
	    this.$el.html(_.template(this.template));
	    return this;
  	}

}

class RepositoriesView extends Backbone.View {

  initialize () {
  	if (window.token == "" || window.token == undefined) {
  		$(location).attr("href", "#")
		return
  	}

    this.template = $('script[name="repositories"]').html();
  }

  render () {
    this.$el.html(_.template(this.template));
    return this;
  }

}

export { HomeView, RepositoriesView, LoginView};