export class Store {
  #state;
  #listeners = new Set();

  constructor(initialState = {}) {
    this.#state = Object.freeze({ ...initialState });
  }

  /** Read-only snapshot of current state. */
  get state() {
    return this.#state;
  }

  setState(partial) {
    this.#state = Object.freeze({ ...this.#state, ...partial });
    this.#listeners.forEach(fn => fn(this.#state));
  }

  subscribe(listener) {
    this.#listeners.add(listener);
    return () => this.#listeners.delete(listener);
  }

  /** Reset to a fresh initial state. */
  reset(initialState = {}) {
    this.setState(initialState);
  }
}
