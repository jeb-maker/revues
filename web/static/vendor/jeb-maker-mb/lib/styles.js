import { css as i } from "lit";
const t = i`
  :host {
    box-sizing: border-box;
    font-family: var(--mb-font-body);
    color: var(--mb-color-fg);
    max-inline-size: 100%;
    overflow-wrap: anywhere;
  }

  :host *,
  :host *::before,
  :host *::after {
    box-sizing: border-box;
  }

  :host([hidden]) {
    display: none !important;
  }

  .control:focus-visible,
  button:focus-visible,
  a:focus-visible,
  select:focus-visible,
  textarea:focus-visible,
  input:focus-visible {
    outline: var(--mb-focus-ring);
    outline-offset: var(--mb-focus-offset);
  }

  @media (prefers-reduced-motion: reduce) {
    :host,
    :host * {
      transition: none !important;
      animation: none !important;
    }
  }
`, a = i`
  .field {
    display: flex;
    flex-direction: column;
    gap: var(--mb-space-1);
    inline-size: 100%;
  }

  .label {
    font-size: var(--mb-font-size-sm);
    font-weight: 600;
    color: var(--mb-color-fg);
  }

  .label.visually-hidden {
    position: absolute;
    inline-size: 1px;
    block-size: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  .hint,
  .error {
    font-size: var(--mb-font-size-sm);
    margin: 0;
  }

  .hint {
    color: var(--mb-color-muted);
  }

  .error {
    color: var(--mb-color-danger);
  }

  .control {
    inline-size: 100%;
    max-inline-size: 100%;
    min-block-size: 2.5rem;
    min-inline-size: 0;
    padding-block: var(--mb-space-2);
    padding-inline: var(--mb-space-3);
    border: 1px solid var(--mb-color-border);
    border-radius: var(--mb-radius-md);
    background: var(--mb-color-surface);
    color: var(--mb-color-fg);
    font: inherit;
    transition:
      border-color var(--mb-transition),
      box-shadow var(--mb-transition);
  }

  .control:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }

  :host([invalid]) .control {
    border-color: var(--mb-color-danger);
  }

  :host([density='compact']) .field {
    gap: 0;
  }

  :host([density='compact']) .control {
    min-block-size: 2.1rem;
    padding-block: 0.2rem;
    padding-inline: var(--mb-space-2);
    font-size: var(--mb-font-size-sm);
  }

  :host([density='compact']) textarea.control {
    min-block-size: 2.1rem;
  }
`;
function s(o, e, r) {
  return o ? { labelText: o, hideVisually: e, controlAriaLabel: "" } : { labelText: "", hideVisually: !1, controlAriaLabel: r };
}
export {
  s as fieldLabelState,
  a as fieldStyles,
  t as sharedStyles
};
//# sourceMappingURL=styles.js.map
