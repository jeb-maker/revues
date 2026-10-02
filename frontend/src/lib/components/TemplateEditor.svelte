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
	<p class="err" role="alert">{error}</p>
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
	.editor {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}
	.head,
	.row {
		display: grid;
		grid-template-columns: 2rem 1.4fr 1fr 1fr 2.5rem auto;
		gap: 0.4rem;
		align-items: center;
	}
	.head {
		font-size: 0.75rem;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		opacity: 0.7;
		padding: 0 0.15rem;
	}
	.row {
		padding: 0.35rem;
		border-radius: 0.4rem;
		background: color-mix(in srgb, var(--mb-color-canvas, #0f172a) 70%, transparent);
		border: 1px solid color-mix(in srgb, #94a3b8 25%, transparent);
	}
	.row.is-dragging {
		opacity: 0.55;
		outline: 1px dashed #5eead4;
	}
	.drag {
		cursor: grab;
		border: none;
		background: transparent;
		color: #94a3b8;
		font-size: 0.9rem;
		padding: 0.2rem;
	}
	.drag:disabled {
		opacity: 0.3;
		cursor: default;
	}
	input[type='text'] {
		width: 100%;
		box-sizing: border-box;
		padding: 0.4rem 0.5rem;
		border-radius: 0.35rem;
		border: 1px solid #334155;
		background: #0b1220;
		color: #f8fafc;
		font: inherit;
	}
	.req {
		display: flex;
		justify-content: center;
	}
	.actions {
		display: flex;
		gap: 0.15rem;
	}
	.actions button,
	.add button {
		border: none;
		background: transparent;
		color: #5eead4;
		cursor: pointer;
		font: inherit;
		padding: 0.2rem 0.35rem;
	}
	.actions button:disabled {
		opacity: 0.35;
		cursor: default;
	}
	.add {
		margin: 0.5rem 0 0;
	}
	.err {
		color: #fca5a5;
		margin: 0 0 0.5rem;
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
