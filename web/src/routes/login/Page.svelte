<script lang="ts">
  import { login, register } from './auth';

  let { onLogin = () => {} } = $props();

  let mode = $state('login');  
  let email = $state('');
  let password = $state('');
  let confirm = $state('');
  let error = $state('');
  let notice = $state('');
  let busy = $state(false);

  const pwBytes = $derived(new TextEncoder().encode(password).length);
  const isRegister = $derived(mode === 'register');

  function switchMode(next: 'login' | 'register') {
    mode = next;
    error = '';
    notice = '';
    confirm = '';
  }

  function validate() {
    if (!isRegister) return '';
    if (pwBytes < 10) return 'Password needs at least 10 characters.';
    if (pwBytes > 72) return 'Password is too long. Keep it under 72 characters.';
    if (password !== confirm) return 'Passwords don’t match.';
    return '';
  }

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = validate();
    notice = '';
    if (error) return;

    busy = true;
    try {
      if (isRegister) {
        await register(email, password);
        switchMode('login');
        notice = 'Account created. Log in to continue.';
        password = '';
      } else {
        const token = await login(email, password);
        onLogin(token);
      }
    } catch (err: any) {
      error = err.message;
    } finally {
      busy = false;
    }
  }
</script>

<main>
  <div class="panel">
    <h1>{isRegister ? 'Create your account' : 'Welcome back'}</h1>

    <div class="tabs" role="tablist">
      <button role="tab" aria-selected={!isRegister} onclick={() => switchMode('login')}>Log in</button>
      <button role="tab" aria-selected={isRegister} onclick={() => switchMode('register')}>Register</button>
    </div>

    <form onsubmit={submit} novalidate>
      <label>
        Email
        <input type="email" bind:value={email} autocomplete="email" required />
      </label>

      <label>
        Password
        <input
          type="password"
          bind:value={password}
          autocomplete={isRegister ? 'new-password' : 'current-password'}
          required
        />
        {#if isRegister}
          <small class:ok={pwBytes >= 10 && pwBytes <= 72}>10–72 characters</small>
        {/if}
      </label>

      {#if isRegister}
        <label>
          Confirm password
          <input type="password" bind:value={confirm} autocomplete="new-password" required />
        </label>
      {/if}

      {#if error}<p class="msg err" role="alert">{error}</p>{/if}
      {#if notice}<p class="msg ok" role="status">{notice}</p>{/if}

      <button type="submit" class="primary" disabled={busy}>
        {busy ? 'Please wait…' : isRegister ? 'Create account' : 'Log in'}
      </button>
    </form>
  </div>
</main>

<style>
  main {
    --bg: var(--surface);
    --accent: var(--teal-700);
    min-height: 100vh;
    display: grid;
    place-items: center;
    padding: 1.5rem;
    background: var(--bg);
    color: var(--ink);
    font-family: 'Inter', system-ui, sans-serif;
  }
  .panel {
    width: 100%;
    max-width: 22rem;
    background: #fff;
    padding: 2rem;
    border-radius: 12px;
    border: 1px solid var(--line);
  }
  h1 { font-size: 1.5rem; margin: 0 0 1.25rem; font-weight: 650; }
  .tabs { display: flex; gap: 0.25rem; margin-bottom: 1.5rem; border-bottom: 1px solid var(--line); }
  .tabs button {
    flex: 1; padding: 0.6rem; background: none; border: none;
    border-bottom: 2px solid transparent; color: var(--muted);
    font: inherit; cursor: pointer; margin-bottom: -1px;
  }
  .tabs button[aria-selected='true'] { color: var(--ink); border-color: var(--accent); font-weight: 600; }
  form { display: grid; gap: 1rem; }
  label { display: grid; gap: 0.35rem; font-size: 0.9rem; font-weight: 500; }
  input {
    padding: 0.65rem 0.75rem; font: inherit; color: inherit;
    border: 1px solid var(--line); border-radius: 8px;
  }
  input:focus-visible, button:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
  small { color: var(--muted); font-weight: 400; }
  small.ok { color: var(--accent); }
  .msg { margin: 0; font-size: 0.9rem; }
  .msg.err { color: var(--err); }
  .msg.ok { color: var(--accent); }
  .primary {
    padding: 0.75rem; font: inherit; font-weight: 600; color: #fff;
    background: var(--accent); border: none; border-radius: 8px; cursor: pointer;
  }
  .primary:disabled { opacity: 0.6; cursor: wait; }
</style>
