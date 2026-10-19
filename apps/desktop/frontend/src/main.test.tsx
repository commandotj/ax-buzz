import { act } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const { render } = vi.hoisted(() => ({ render: vi.fn() }));

vi.mock("react-dom/client", () => ({
  default: { createRoot: vi.fn(() => ({ render })) },
  createRoot: vi.fn(() => ({ render })),
}));
vi.mock("./App", () => ({ default: () => <div>Buzz app</div> }));

describe("frontend bootstrap", () => {
  beforeEach(() => {
    render.mockClear();
    document.body.innerHTML = '<div id="root"></div>';
  });

  it("mounts the application root", async () => {
    await act(async () => { await import("./main"); });
    expect(render).toHaveBeenCalledOnce();
  });
});
