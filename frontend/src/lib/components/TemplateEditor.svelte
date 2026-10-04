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

<style>
	.editor {
		display: flex;
		flex-direction: column;
		gap: var(--mb-space-1);
	}
	.head,
	.row {
		display: grid;
		grid-template-columns: 2rem minmax(0, 1.4fr) minmax(0, 1fr) minmax(0, 1fr) 2.5rem 5.75rem;
		gap: var(--mb-space-2);
		align-items: center;
		padding: var(--mb-space-1);
		border: 1px solid transparent;
		box-sizing: border-box;
	}
	.head {
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--mb-color-muted);
	}
	.head .col-field {
		padding-inline: var(--mb-space-2);
	}
	.head .col-req {
		text-align: center;
	}
	.row {
		position: relative;
		background: var(--mb-color-surface);
		border-color: var(--mb-color-border);
		border-radius: var(--mb-radius-sm);
	}
	.row.is-dragging {
		opacity: 0.55;
		outline: 1px dashed var(--mb-color-accent);
	}
	.drag {
		padding: var(--mb-space-1);
		border: 0;
		background: transparent;
		color: var(--mb-color-muted);
		cursor: grab;
	}
	.drag:disabled {
		opacity: 0.3;
		cursor: default;
	}
	input[type='text'],
	.help textarea {
		width: 100%;
		min-width: 0;
		padding: var(--mb-space-2);
		border: 1px solid var(--mb-color-border);
		border-radius: var(--mb-radius-sm);
		background: var(--mb-color-bg);
		color: var(--mb-color-fg);
		font: inherit;
		line-height: 1.25;
	}
	.help {
		position: relative;
		min-width: 0;
	}
	.help textarea {
		display: block;
		height: 2.375rem;
		max-height: 2.375rem;
		resize: none;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}
	.help textarea:focus {
		position: absolute;
		left: 0;
		right: 0;
		top: 0;
		z-index: 4;
		max-height: none;
		min-height: 7rem;
		height: auto;
		white-space: pre-wrap;
		overflow: auto;
		box-shadow: 0 0.35rem 0.75rem var(--mb-color-border);
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
		border: 0;
		background: transparent;
		color: var(--mb-color-accent);
		font: inherit;
		cursor: pointer;
	}
	.actions button:disabled {
		opacity: 0.35;
	}
	.add {
		margin: var(--mb-space-2) 0 0;
	}
	@media (max-width: 720px) {
		.head {
			display: none;
		}
		.row {
			grid-template-columns: 2rem 1fr;
		}
		.row > :not(.drag) {
			grid-column: 2;
		}
		.help textarea {
			position: static;
			max-height: none;
			min-height: 4rem;
			height: auto;
			white-space: pre-wrap;
			overflow: auto;
		}
		.help textarea:focus {
			position: static;
			box-shadow: none;
		}
	}
</style>
