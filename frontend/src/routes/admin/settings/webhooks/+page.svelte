<script lang="ts">
	import { onMount } from 'svelte';
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
	import { session } from '$lib/auth/session';
	import { formatDeliveryState } from '$lib/i18n/labels';
	import { inputValue } from '$lib/mb';

	const csrf = session().csrf_token;

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
		} catch (err) {
			error = err instanceof Error ? err.message : 'Test impossible.';
		}
		await refreshDeliveries().catch(() => undefined);
	}

	async function onClear() {
		if (!confirm('Effacer la configuration des webhooks ?')) return;
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
			message = 'File de livraisons traitée.';
			await refreshDeliveries();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Traitement impossible.';
		}
	}

	async function onRetry(id: number) {
		error = '';
		message = '';
		try {
			await retryAdminWebhookDelivery(id, csrf);
			message = `Livraison #${id} relancée.`;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Relance impossible.';
		}
		await refreshDeliveries().catch(() => undefined);
	}
</script>

<svelte:head>
	<title>Webhooks — Revues</title>
</svelte:head>

	<header class="page-header">
		<p class="crumbs">
			<a href="/admin">Administration</a> · <a href="/admin/integrations">Intégrations</a> · Webhooks
		</p>
		<h1>Webhooks sortants</h1>
		<p class="lede">
			Destinations qui reçoivent les événements Revues. Chaque envoi est signé avec le secret
			pour que le destinataire puisse vérifier l’origine.
		</p>
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}
	{#if message}
		<mb-alert variant="success">{message}</mb-alert>
	{/if}

		{#if loading}
			<p class="loading">Chargement…</p>
		{:else}
			<form class="stack-form" onsubmit={onSave}>
				<mb-textarea
					label="URLs"
					hint="Une URL par ligne."
					rows="4"
					required
					placeholder="https://hooks.example.com/revues"
					value={urlsText}
					oninput={(e) => (urlsText = inputValue(e))}
				></mb-textarea>
				<mb-input
					label="Secret de signature"
					type="password"
					autocomplete="new-password"
					hint={hasSecret ? 'Un secret est enregistré ; laissez vide pour le conserver.' : undefined}
					value={secret}
					oninput={(e) => (secret = inputValue(e))}
				></mb-input>
				<mb-checkbox
					checked={reviewCompleted}
					onmb-change={(e) => (reviewCompleted = !!e.detail.checked)}
				>
					Quand une revue est clôturée
				</mb-checkbox>
				<mb-checkbox
					checked={reviewItemNok}
					onmb-change={(e) => (reviewItemNok = !!e.detail.checked)}
				>
					Quand un point est marqué non validé
				</mb-checkbox>
				<p class="actions">
					<mb-button type="submit" variant="primary" loading={saving}>
						{saving ? 'Enregistrement…' : 'Enregistrer'}
					</mb-button>
					{#if configured}
						<mb-button type="button" variant="secondary" onclick={onTest}>Envoyer un test</mb-button>
						<mb-button type="button" variant="danger" onclick={onClear}>Effacer</mb-button>
					{/if}
				</p>
			</form>

			<section class="section">
				<div class="actions">
					<h2>File de livraisons</h2>
					<mb-button type="button" variant="ghost" size="sm" onclick={onDrain}>
						Traiter maintenant
					</mb-button>
				</div>
				{#if deliveries.length === 0}
					<p class="muted">Aucune livraison pour cette organisation.</p>
				{:else}
					<div class="table-scroll">
						<table>
							<thead>
								<tr>
									<th scope="col">ID</th>
									<th scope="col">Événement</th>
									<th scope="col">État</th>
									<th scope="col">Tentatives</th>
									<th scope="col">HTTP</th>
									<th scope="col"><span class="sr-only">Actions</span></th>
								</tr>
							</thead>
							<tbody>
								{#each deliveries as d (d.id)}
									<tr>
										<td>{d.id}</td>
										<td>
											<code>{d.event_type}</code>
											<span class="desc url">{d.url}</span>
										</td>
										<td>{formatDeliveryState(d.state)}</td>
										<td>{d.attempts}</td>
										<td>{d.status_code ?? '—'}</td>
										<td>
											{#if d.state === 'pending' || d.state === 'poison'}
												<mb-button
													type="button"
													variant="ghost"
													size="sm"
													icon-only
													aria-label="Relancer"
													title="Relancer"
													onclick={() => onRetry(d.id)}
												>
													<svg class="icon" viewBox="0 0 24 24" aria-hidden="true"
														><path d="M21 12a9 9 0 1 1-2.64-6.36M21 3v6h-6" /></svg
													>
												</mb-button>
											{/if}
										</td>
									</tr>
									{#if d.last_error}
										<tr>
											<td colspan="6" class="field-error">{d.last_error}</td>
										</tr>
									{/if}
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</section>
		{/if}

<style>
	.url {
		word-break: break-all;
	}
	.section .actions > h2 {
		margin: 0;
		flex: 1;
	}
</style>
