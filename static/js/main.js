app = angular.module('revco', ["ngRoute"]);

app.config(['$routeProvider',
	function($routeProvider) {
		$routeProvider.
			when('/', {
				templateUrl: '/partials/home.html',
				controller: 'MainCtrl'
			}).
			when('/me/view', {
				templateUrl: 'partials/my_projects.html',
				controller: 'MyProjectsCtrl'
			}).
			when('/view/:project', {
				templateUrl: '/partials/view_project.html'
			}).
			when('/explore', {
				templateUrl: '/partials/explore.html'
			}).
			when('/make', {
				templateUrl: '/partials/make.html',
				controller: 'MakeCtrl'
			}).
			when('/auth/:method', {
				templateUrl: '/partials/login.html',
				controller: 'AuthCtrl'
			}).
			when('/login', {
				redirectTo: '/auth/login'
			}).
			when('/register', {
				redirectTo: '/auth/register'
			}).
			otherwise({
				redirectTo: '/'
			});

	}]);

app.service('User', function($rootScope, $http, $location) {
	var service = {
		username: undefined, 
		token: localStorage.token,
		changeUsername: function(username) {
			service.username = username;
			$rootScope.$broadcast('user.update');
		},
		changeToken: function(token) {
			service.token = token;
			$rootScope.$broadcast('user.update');
		},
		loggedIn: function() {
			console.log(service.token);
			return service.token !== undefined && service.token != "undefined";
		},
		auth: function(method, username, password, email, error) {
			console.log(username, password, email);
			error = error || function(m) {console.log(m);};
			var url = '/api/user/' + method + "?username=" + username + "&password=" + password + "&email=" + email;
			console.log(url);
			$http.post(url).
				success(function(data) {
					console.log(data);
					if (data.status.error) {
					 	error(data.status.message);
						return;	
					}
					
					if (data.data.username !== undefined) {
						error("Login with your fresh account!");
						$location.path('/login');
						return;
					}

					localStorage.token = data.data;
					service.changeToken(data.data);
					$location.path("/");
					
					service.info();
				}).
				error(function(data) {
					console.log(data);
				});
		},
		info: function() {
			console.log("in info", service.token);
			$http.get('/api/user?access_token=' + service.token).
				success(function(data) {
					console.log(data);
					service.token = undefined;
					if (data.status.error) {
						$location.path("/login");
						return; 
					}
					service.changeToken(localStorage.token);
					service.changeUsername(data.data.username);
				}).
				error(function(data) {console.log("Unable to get user :/");});
		}
	};

	return service;
});

app.controller('MainCtrl', function($scope, User) {
	$scope.$on('user.update', function(event) {
		$scope.message = User.username;
	});
	$scope.message = User.username;

	$scope.change = function() {
		User.changeUsername("boo I liked that username!");
	};
});

app.controller('AuthCtrl', function($scope, $routeParams, $http, $location, User) {
	if(User.loggedIn()) {
		$location.path("#/");
		return;
	}

	$scope.method = $routeParams.method;
	$scope.email = " ";

	$scope.other = function() {
		return $scope.method == "login" ? "register" : "login";
	};

	$scope.go = function() {
		console.log($scope.email);
		username = $scope.username;
		password = $scope.password;
		email = $scope.email;

		if (!username) {
			$scope.setError("You are *going* to need a username, trust me");
			return;
		}

		if (!password) {
			$scope.setError("You need a password to login!");
			return;
		}

		if (!email && $scope.method == "register") {
			$scope.setError("You need an email!");
			return;
		}
		User.auth($scope.method, $scope.username, $scope.password, $scope.email, $scope.setError);
	};

	$scope.setError = function(error) {
		$scope.error = error;
	};
});

app.controller('MakeCtrl', function($scope, $http, User) {
	$scope.make = function() {
		name = $scope.name;
		url = $scope.url;

		if(!name) {
			return;
		}

		if(!url) {
			return;
		}
		
		var urlParts = url.split('/');
		var urlHost = urlParts[0];
		var urlUser = urlParts[1];
		var urlProject = urlParts[2];
		
		var url = '/api/repo/' + urlHost + '/' + urlUser + '/' + urlProject + '?access_token=' + User.token;
		console.log(url);
		$http.post(url)
			.success(function(data) {
				if(data.status.error) {
					console.log(data.status);
				}
			})
			.error(function() {
			
			});
	};
});

app.controller('MyProjectsCtrl', function($scope, $http, User) {
	$http.get("/api/user/repositories?access_token=" + User.token).
		success(function(data) {
			if (data.status.error) {
				console.log("Error in repositories :/");
				return;
			}

			$scope.repositories = data.data || [];
		}).
		error(function() {
			console.log("Error fetching repositories!");
		});
});

app.controller('HeaderCtrl', function($scope, $location, $http, User) {
	$scope.$on('user.update', function(event) {
		$scope.user = User;
	});

	if(!User.loggedIn()) {
		$location.path("/login");
		return;
	}

	User.info();

	$scope.user = User;
});
