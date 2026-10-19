import { describe, expect, it } from "vitest";
import { stageScale } from "./components/Stage";

describe("stageScale", () => {
  it("fits 16:9 inside the window", () => {
    expect(stageScale(1600, 900)).toBe(1);
    expect(stageScale(1920, 1080)).toBe(1.2);
    expect(stageScale(1280, 1024)).toBe(0.8); // width-bound, letterboxed top and bottom
    expect(stageScale(3000, 900)).toBe(1); // height-bound
    expect(stageScale(6400, 3600)).toBe(2); // capped
  });
});
