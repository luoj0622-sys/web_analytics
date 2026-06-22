export function createTracker(options) {
  const env = options.env ?? globalThis;
  const context = createContext(options, env);

  async function send(type, fields = {}) {
    const payload = {
      type,
      site_id: context.siteId,
      public_key: context.publicKey,
      occurred_at: env.now ? env.now().toISOString() : new Date().toISOString(),
      visitor: {
        id: context.visitorId,
        session_id: context.sessionId
      },
      page: currentPage(env),
      campaign: campaignFromLocation(env.location),
      device: {
        user_agent: env.navigator?.userAgent ?? '',
        language: env.navigator?.language ?? ''
      },
      ...fields
    };

    await env.fetch(context.collectUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(payload),
      keepalive: true
    });
  }

  return {
    context,
    trackPageView: () => send('page_view'),
    heartbeat: () => send('heartbeat'),
    track: (name, properties = {}) => send('custom', { name, properties })
  };
}

function createContext(options, env) {
  const siteId = options.siteId;
  const publicKey = options.publicKey;
  const collectUrl = options.collectUrl;
  const visitorKey = `wa:${siteId}:visitor_id`;
  const sessionKey = `wa:${siteId}:session_id`;

  const visitorId = getOrSet(env.localStorage, visitorKey, () => `v_${randomID(env)}`);
  const sessionId = getOrSet(env.localStorage, sessionKey, () => `s_${randomID(env)}`);

  return {
    siteId,
    publicKey,
    collectUrl,
    visitorId,
    sessionId
  };
}

function getOrSet(storage, key, createValue) {
  const existing = storage?.getItem(key);
  if (existing) {
    return existing;
  }
  const value = createValue();
  storage?.setItem(key, value);
  return value;
}

function randomID(env) {
  if (typeof env.random === 'function') {
    return env.random();
  }
  const crypto = env.crypto;
  if (crypto?.randomUUID) {
    return crypto.randomUUID().replaceAll('-', '');
  }
  return Math.random().toString(36).slice(2);
}

function currentPage(env) {
  return {
    url: env.location?.href ?? '',
    path: env.location?.pathname ?? '',
    title: env.document?.title ?? '',
    referrer: env.document?.referrer ?? ''
  };
}

function campaignFromLocation(location) {
  const params = new URLSearchParams(location?.search ?? '');
  return {
    source: params.get('utm_source') ?? '',
    medium: params.get('utm_medium') ?? '',
    campaign: params.get('utm_campaign') ?? '',
    term: params.get('utm_term') ?? '',
    content: params.get('utm_content') ?? ''
  };
}
