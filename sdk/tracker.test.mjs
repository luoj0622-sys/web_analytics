import test from 'node:test';
import assert from 'node:assert/strict';
import { createTracker } from './tracker.mjs';

test('bootstrap creates visitor and session identifiers', () => {
  const env = createEnv();
  const tracker = createTracker({
    siteId: 'site_1',
    publicKey: 'pk_1',
    collectUrl: 'https://analytics.example.com/collect',
    env
  });

  assert.equal(tracker.context.siteId, 'site_1');
  assert.ok(tracker.context.visitorId.startsWith('v_'));
  assert.ok(tracker.context.sessionId.startsWith('s_'));
  assert.equal(env.localStorage.getItem('wa:site_1:visitor_id'), tracker.context.visitorId);
});

test('trackPageView sends common analytics fields', async () => {
  const env = createEnv();
  const tracker = createTracker({
    siteId: 'site_1',
    publicKey: 'pk_1',
    collectUrl: 'https://analytics.example.com/collect',
    env
  });

  await tracker.trackPageView();

  assert.equal(env.requests.length, 1);
  const body = JSON.parse(env.requests[0].body);
  assert.equal(body.type, 'page_view');
  assert.equal(body.site_id, 'site_1');
  assert.equal(body.public_key, 'pk_1');
  assert.equal(body.page.url, 'https://example.com/pricing?utm_source=newsletter&utm_medium=email');
  assert.equal(body.campaign.source, 'newsletter');
  assert.equal(body.campaign.medium, 'email');
  assert.equal(body.visitor.id, tracker.context.visitorId);
  assert.equal(body.visitor.session_id, tracker.context.sessionId);
});

test('heartbeat sends heartbeat event', async () => {
  const env = createEnv();
  const tracker = createTracker({ siteId: 'site_1', publicKey: 'pk_1', collectUrl: '/collect', env });

  await tracker.heartbeat();

  const body = JSON.parse(env.requests[0].body);
  assert.equal(body.type, 'heartbeat');
});

test('custom event sends name and properties', async () => {
  const env = createEnv();
  const tracker = createTracker({ siteId: 'site_1', publicKey: 'pk_1', collectUrl: '/collect', env });

  await tracker.track('signup', { plan: 'pro' });

  const body = JSON.parse(env.requests[0].body);
  assert.equal(body.type, 'custom');
  assert.equal(body.name, 'signup');
  assert.deepEqual(body.properties, { plan: 'pro' });
});

function createEnv() {
  const storage = new Map();
  const requests = [];
  return {
    requests,
    document: {
      title: 'Pricing',
      referrer: 'https://search.example'
    },
    location: {
      href: 'https://example.com/pricing?utm_source=newsletter&utm_medium=email',
      pathname: '/pricing',
      search: '?utm_source=newsletter&utm_medium=email'
    },
    navigator: {
      userAgent: 'Mozilla/5.0 Test',
      language: 'en-US'
    },
    localStorage: {
      getItem: key => storage.get(key) ?? null,
      setItem: (key, value) => storage.set(key, value)
    },
    fetch: async (url, options) => {
      requests.push({ url, ...options });
      return { ok: true };
    },
    now: () => new Date('2026-06-22T10:30:00Z'),
    random: () => 'abc123'
  };
}
