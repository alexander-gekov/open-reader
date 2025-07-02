export default defineNuxtRouteMiddleware((to) => {
  const unreleasedRoutes = ["/reader", "/settings"];

  if (unreleasedRoutes.includes(to.path)) {
    return navigateTo("/");
  }
});
