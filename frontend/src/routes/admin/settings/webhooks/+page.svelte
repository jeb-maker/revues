<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bootstrap } from '$lib/api/auth';
	import {
		deleteAdminWebhooks,
		drainAdminWebhookDeliveries,
		getAdminWebhooks,
		listAdminWebhookDeliveries,
		postAdminWebhooksTest,
		putAdminWebhooks,
		retryAdminWebhookDelivery,
		type WebhookDelivery,
		type WebhookSettings
	} from '$lib/api/admin';

	let csrf = $state('');
	let error = $state('');
	let message = $state('');
	let loading = $state(true);
	let saving = $state(false);

	let urlsText = $state('');
	let secret = $state('');
	let hasSecret = $state(false);
	let configured = $state(false);
	let reviewCompleted = $state(true);
	let reviewItemNok = $state(false);
	let deliveries = $state<WebhookDelivery[]>([]);

	function apply(s: WebhookSettings) {
		urlsText = (s.urls ?? []).join('\n');
		hasSecret = s.has_secret;
		configured = s.configured;
		reviewCompleted = s.review_completed;
		reviewItemNok = s.review_item_nok;
		secret = '';
	}

	async function refreshDeliveries() {
		deliveries = (await listAdminWebhookDeliveries(csrf)).deliveries ?? [];
	}

	onMount(async () => {
		try {
			const boot = await bootstrap();
			if (!boot.authenticated) {
				await goto('/login');
				return;
			}
			csrf = boot.csrf_token;
			apply(await getAdminWebhooks(csrf));
			await refreshDeliveries();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Webhooks indisponibles.';
		} finally {
			loading = false;
		}
	});

	async function onSave(e: Event) {
		e.preventDefault();
		saving = true;
		error = '';
		message = '';
		try {
			const urls = urlsText
				.split(/\n|,/)
				.map((u) => u.trim())
				.filter(Boolean);
			apply(
				await putAdminWebhooks(
					{
						urls,
						review_completed: reviewCompleted,
						review_item_nok: reviewItemNok,
						...(secret ? { secret } : { secret: '' })
					},
					csrf
				)
			);
			message = 'Configuration webhooks enregistrée.';
			await refreshDeliveries();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Enregistrement impossible.';
		} finally {
			saving = false;
		}
	}

	async function onTest() {
		error = '';
		message = '';
		try {
			await postAdminWebhooksTest(csrf);
			message = 'Événement webhook.test envoyé.';
			await refreshDeliveries();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Test impossible.';
			await refreshDeliveries();
		}
	}

	async function onClear() {
		error = '';
		message = '';
		try {
			await deleteAdminWebhooks(csrf);
			apply({
				configured: false,
				enabled: false,
				urls: [],
				has_secret: false,
				review_completed: true,
				review_item_nok: false
			});
			message = 'Configuration webhooks effacée.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Suppression impossible.';
		}
	}

	async function onDrain() {
		error = '';
		message = '';
		try {
			await drainAdminWebhookDeliveries(csrf);
			message = 'Drain exécuté.';
			await refreshDeliveries();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Drain impossible.';
		}
	}

	async function onRetry(id: number) {
		error = '';
		message = '';
		try {
			await retryAdminWebhookDelivery(id, csrf);
			message = `Livraison #${id} relancée.`;
			await refreshDeliveries();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Retry impossible.';
			await refreshDeliveries();
		}
	}
</script>

<svelte:head>
	<title>Webhooks — Revues</title>
</svelte:head>

<main class="admin-page">
	<header>
		<p class="brand"><a href="/">Revues</a></p>
		<p class="crumbs">
			<a href="/">Accueil</a> · <a href="/admin/integrations">Intégrations</a> · Webhooks
		</p>
		<h1>Webhooks sortants</h1>
		<p class="lede">
			URLs + secret HMAC chiffré. Signature
			<code>X-Revues-Signature: sha256=…</code>. Anti-SSRF à chaque tentative.
		</p>
	</header>

	{#if error}
		<p class="err" role="alert">{error}</p>
	{/if}
	{#if message}
		<p class="ok" role="status">{message}</p>
	{/if}

	{#if loading}
		<p class="muted">Chargement…</p>
	{:else}
		<form class="form" onsubmit={onSave}>
			<label>
				URLs (une par ligne)
				<textarea rows="4" bind:value={urlsText} required placeholder="https://hooks.example.com/revues"></textarea>
			</label>
			<label>
				Secret HMAC
				<input
					type="password"
					bind:value={secret}
					placeholder={hasSecret ? '•••••••• (laisser vide pour conserver)' : ''}
					autocomplete="new-password"
				/>
			</label>
			<label class="check">
				<input type="checkbox" bind:checked={reviewCompleted} />
				Événement <code>review.completed</code>
			</label>
			<label class="check">
				<input type="checkbox" bind:checked={reviewItemNok} />
				Événement <code>review.item.nok</code>
			</label>
			<div class="actions">
				<button type="submit" disabled={saving}>{saving ? 'Enregistrement…' : 'Enregistrer'}</button>
				{#if configured}
					<button type="button" class="ghost" onclick={onClear}>Effacer</button>
					<button type="button" class="ghost" disabled={!configured} onclick={onTest}>Envoyer un test</button>
				{/if}
			</div>
		</form>

		<section class="deliveries">
			<div class="deliveries-head">
				<h2>File de livraisons</h2>
				<button type="button" class="ghost" onclick={onDrain}>Drain maintenant</button>
			</div>
			{#if deliveries.length === 0}
				<p class="muted">Aucune livraison pour cette organisation.</p>
			{:else}
				<table>
					<thead>
						<tr>
							<th>ID</th>
							<th>Événement</th>
							<th>État</th>
							<th>Tentatives</th>
							<th>HTTP</th>
							<th></th>
						</tr>
					</thead>
					<tbody>
						{#each deliveries as d (d.id)}
							<tr>
								<td>{d.id}</td>
								<td>
									<code>{d.event_type}</code>
									<small>{d.url}</small>
								</td>
								<td>{d.state}</td>
								<td>{d.attempts}</td>
								<td>{d.status_code ?? '—'}</td>
								<td>
									{#if d.state === 'pending' || d.state === 'poison'}
										<button type="button" class="linkish" onclick={() => onRetry(d.id)}>Retry</button>
									{/if}
								</td>
							</tr>
							{#if d.last_error}
								<tr class="err-row">
									<td colspan="6">{d.last_error}</td>
								</tr>
							{/if}
						{/each}
					</tbody>
				</table>
			{/if}
		</section>
	{/if}
</main>

<style>
	.form,
	.deliveries {
		display: flex;
		flex-direction: column;
		gap: 0.85rem;
	}
	.deliveries {
		margin-top: 2rem;
		padding-top: 1.25rem;
		border-top: 1px solid #334155;
	}
	.deliveries-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		flex-wrap: wrap;
	}
	.deliveries h2 {
		margin: 0;
		font-size: 1rem;
		color: #99f6e4;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		font-size: 0.9rem;
		color: #cbd5e1;
	}
	label.check {
		flex-direction: row;
		align-items: center;
		gap: 0.5rem;
	}
	input:not([type='checkbox']),
	textarea {
		padding: 0.55rem 0.65rem;
		border-radius: 0.4rem;
		border: 1px solid #334155;
		background: #0f172a;
		color: inherit;
		font: inherit;
	}
	.actions {
		display: flex;
		gap: 0.75rem;
		flex-wrap: wrap;
	}
	button {
		align-self: flex-start;
		padding: 0.6rem 1rem;
		border: none;
		border-radius: 0.4rem;
		background: #0f766e;
		color: #ecfdf5;
		font-weight: 600;
		cursor: pointer;
		font: inherit;
	}
	button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	button.ghost {
		background: transparent;
		border: 1px solid #475569;
		color: #cbd5e1;
	}
	button.linkish {
		background: transparent;
		color: #5eead4;
		padding: 0.2rem 0.4rem;
		font-weight: 500;
		font-size: 0.85rem;
	}
	table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.85rem;
	}
	th,
	td {
		text-align: left;
		padding: 0.45rem 0.35rem;
		border-bottom: 1px solid #1e293b;
		vertical-align: top;
	}
	td small {
		display: block;
		color: #94a3b8;
		word-break: break-all;
	}
	.err-row td {
		color: #fca5a5;
		font-size: 0.8rem;
		border-bottom: 1px solid #334155;
	}
	code {
		font-size: 0.85em;
	}
</style>
