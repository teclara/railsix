/** Thin wrapper around Umami's track function. No-ops if Umami isn't loaded. */
export function track(event: string, data?: Record<string, string | number | boolean>) {
	window.umami?.track(event, data);
}
