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
		<span>Case</span>
		<span>Catégorie</span>
		<span>Aide</span>
		<span>Oblig.</span>
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
			<input type="text" bind:value={item.label} aria-label="Case" placeholder="Libellé" />
			<input
				type="text"
				bind:value={item.section}
				aria-label="Catégorie"
				placeholder="Catégorie"
			/>
			<input type="text" bind:value={item.help_text} aria-label="Aide" placeholder="Aide" />
			<label class="req">
				<input type="checkbox" bind:checked={item.required} />
				<span class="visually-hidden">Obligatoire</span>
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

<style>
	/* Éditeur de modèle : grille + DnD natifs (exception actée, decisions.md) — tokens mb. */
	.editor {
		display: flex;
		flex-direction: column;
		gap: var(--mb-space-1);
	}
	.head,
	.row {
		display: grid;
		grid-template-columns: 2rem 1.4fr 1fr 1fr 2.5rem auto;
		gap: var(--mb-space-2);
		align-items: center;
	}
	.head {
		padding: 0 var(--mb-space-1);
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--mb-color-muted);
	}
	.row {
		padding: var(--mb-space-1);
		background: var(--mb-color-surface);
		border: 1px solid var(--mb-color-border);
		border-radius: var(--mb-radius-sm);
	}
	.row.is-dragging {
		opacity: 0.55;
		outline: 1px dashed var(--mb-color-accent);
	}
	.drag {
		padding: var(--mb-space-1);
		border: none;
		background: transparent;
		color: var(--mb-color-muted);
		font-size: 0.9rem;
		cursor: grab;
	}
	.drag:disabled {
		opacity: 0.3;
		cursor: default;
	}
	input[type='text'] {
		width: 100%;
		padding: var(--mb-space-2);
		border: 1px solid var(--mb-color-border);
		border-radius: var(--mb-radius-sm);
		background: var(--mb-color-bg);
		color: var(--mb-color-fg);
		font: inherit;
	}
	input[type='checkbox'] {
		accent-color: var(--mb-color-accent);
	}
	.req {
		display: flex;
		justify-content: center;
	}
	.actions {
		display: flex;
		gap: var(--mb-space-1);
	}
	.actions button,
	.add button {
		padding: var(--mb-space-1) var(--mb-space-2);
		border: none;
		background: transparent;
		color: var(--mb-color-accent);
		font: inherit;
		cursor: pointer;
	}
	.actions button:disabled {
		opacity: 0.35;
		cursor: default;
	}
	.add {
		margin: var(--mb-space-2) 0 0;
	}
	.visually-hidden {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		border: 0;
	}
	@media (max-width: 720px) {
		.head {
			display: none;
		}
		.row {
			grid-template-columns: 2rem 1fr;
		}
		.row > input,
		.row > .req,
		.row > .actions {
			grid-column: 2;
		}
	}
</style>
