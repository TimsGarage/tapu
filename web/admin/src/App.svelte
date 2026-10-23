<script>
	import AccountPage from "./pages/AccountPage.svelte";
	import AdminPage from "./pages/AdminPage.svelte";
	import LoginPage from "./pages/LoginPage.svelte";

	const pathname = window.location.pathname;

	const CurrentPage = (() => {
		if (pathname === "/login") {
			return LoginPage;
		}

		if (pathname === "/account") {
			return AccountPage;
		}

		if (pathname === "/admin" || pathname.startsWith("/admin/")) {
			return AdminPage;
		}

		return AccountPage;
	})();
</script>

<CurrentPage />
