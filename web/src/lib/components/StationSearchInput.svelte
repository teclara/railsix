<script lang="ts">
	import type { Stop } from '$lib/api';

	let {
		stops,
		placeholder = 'Search stations...',
		value = $bindable(''),
		onSelect
	}: {
		stops: Stop[];
		placeholder?: string;
		value?: string;
		onSelect: (stop: Stop) => void;
	} = $props();

	const sortedStops = $derived([...stops].sort((a, b) => a.name.localeCompare(b.name)));
	let results = $state<Stop[]>([]);
	let showDropdown = $state(false);

	function search(q: string) {
		value = q;
		if (q.length === 0) {
			results = sortedStops;
		} else {
			const lower = q.toLowerCase();
			results = sortedStops.filter((s) => s.name.toLowerCase().includes(lower));
		}
		showDropdown = results.length > 0;
	}

	function onFocus() {
		if (value.length === 0) {
			results = sortedStops;
		} else {
			const lower = value.toLowerCase();
			results = sortedStops.filter((s) => s.name.toLowerCase().includes(lower));
		}
		showDropdown = results.length > 0;
	}

	let tappingDropdown = false;

	function select(stop: Stop) {
		tappingDropdown = false;
		value = stop.name;
		results = [];
		showDropdown = false;
		onSelect(stop);
	}

	function onBlur() {
		setTimeout(() => {
			if (!tappingDropdown) showDropdown = false;
		}, 200);
	}
</script>

<div class="station-search-input relative">
	<input
		type="text"
		class="w-full bg-[#1e1e1e] text-white font-mono px-3 py-2 rounded border border-[#333] focus:border-amber-400 focus:outline-none"
		style="font-size: 16px;"
		{placeholder}
		{value}
		oninput={(e) => search((e.target as HTMLInputElement).value)}
		onfocus={onFocus}
		onclick={onFocus}
		onblur={onBlur}
		autocomplete="off"
	/>
	{#if showDropdown}
		<ul
			class="dropdown absolute z-50 w-full mt-1 bg-[#1e1e1e] border border-[#333] rounded shadow-lg max-h-64 overflow-y-auto"
			role="listbox"
			onpointerdown={() => (tappingDropdown = true)}
		>
			{#each results as stop}
				<li role="option" aria-selected="false">
					<button
						class="w-full text-left px-3 py-2 font-mono text-white hover:bg-[#2a2a2a] focus:bg-[#2a2a2a]"
						style="font-size: 16px;"
						onclick={() => select(stop)}
					>
						{stop.name}
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>
