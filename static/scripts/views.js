class HomeView extends Backbone.View {

  initialize () {
    this.template = $('script[name="home"]').html();
  }

  render () {
    this.$el.html(_.template(this.template));
    return this;
  }

}

class RepositoriesView extends Backbone.View {

  initialize () {
    this.template = $('script[name="repositories"]').html();
  }

  render () {
    this.$el.html(_.template(this.template));
    return this;
  }

}

export { HomeView, RepositoriesView };