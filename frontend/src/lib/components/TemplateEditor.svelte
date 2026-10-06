<script lang="ts">
	import type { EditorItem } from './templateEditorModel';
	import { emptyEditorItem } from './templateEditorModel';

	let {
		items = $bindable([] as EditorItem[]),
		error = ''
	}: {
		items?: EditorItem[];
		error?: string;
	} = $props();

	let dragIndex = $state<number | null>(null);

	function addRow() {
		const prev = items[items.length - 1];
		items = [...items, emptyEditorItem(prev?.section ?? '')];
	}

	function removeRow(i: number) {
		if (items.length <= 1) return;
		items = items.filter((_, idx) => idx !== i);
	}

	function move(i: number, delta: number) {
		const j = i + delta;
		if (j < 0 || j >= items.length) return;
		const next = [...items];
		const [row] = next.splice(i, 1);
		next.splice(j, 0, row);
		items = next;
	}

	function onDragStart(i: number, e: DragEvent) {
		dragIndex = i;
		e.dataTransfer?.setData('text/plain', String(i));
		if (e.dataTransfer) e.dataTransfer.effectAllowed = 'move';
	}

	function onDragOver(i: number, e: DragEvent) {
		e.preventDefault();
		if (dragIndex === null || dragIndex === i) return;
		const next = [...items];
		const [row] = next.splice(dragIndex, 1);
		next.splice(i, 0, row);
		items = next;
		dragIndex = i;
	}

	function onDragEnd() {
		dragIndex = null;
	}
</script>

{#if error}
	<p class="field-error" role="alert">{error}</p>
{/if}

<div class="editor" class:dragging={dragIndex !== null}>
	<div class="head" aria-hidden="true">
		<span class="col-drag"></span>
		<span class="col-field">Titre</span>
		<span class="col-field">Catégorie</span>
		<span class="col-field">Explication</span>
		<span class="col-req">Oblig.</span>
		<span></span>
	</div>
	{#each items as item, i (item.key)}
		<div
			class="row"
			class:is-dragging={dragIndex === i}
			ondragover={(e) => onDragOver(i, e)}
			role="listitem"
		>
			<button
				type="button"
				class="drag"
				draggable="true"
				aria-label="Déplacer cette case"
				title="Glisser pour réordonner"
				ondragstart={(e) => onDragStart(i, e)}
				ondragend={onDragEnd}
				disabled={items.length <= 1}
			>
				⋮⋮
			</button>
			<input type="text" bind:value={item.label} aria-label="Titre" placeholder="Titre" />
			<input
				type="text"
				bind:value={item.section}
				aria-label="Catégorie"
				placeholder="Catégorie"
			/>
			<div class="help">
				<textarea
					bind:value={item.help_text}
					aria-label="Explication"
					placeholder="Explication"
					rows="1"
					title={item.help_text || 'Explication'}
				></textarea>
			</div>
			<label class="req">
				<input type="checkbox" bind:checked={item.required} />
				<span class="sr-only">Obligatoire</span>
			</label>
			<div class="actions">
				<button type="button" aria-label="Monter" disabled={i === 0} onclick={() => move(i, -1)}
					>↑</button
				>
				<button
					type="button"
					aria-label="Descendre"
					disabled={i === items.length - 1}
					onclick={() => move(i, 1)}>↓</button
				>
				<button
					type="button"
					aria-label="Supprimer"
					disabled={items.length <= 1}
					onclick={() => removeRow(i)}>×</button
				>
			</div>
		</div>
	{/each}
	<p class="add">
		<button type="button" onclick={addRow}>+ Ajouter une case</button>
	</p>
</div>
