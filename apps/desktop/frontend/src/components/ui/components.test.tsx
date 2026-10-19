import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Badge } from "./badge";
import { Button } from "./button";
import {
  Card, CardAction, CardContent, CardDescription, CardFooter, CardHeader, CardTitle,
} from "./card";
import { Separator } from "./separator";
import { Skeleton } from "./skeleton";

describe("shared UI components", () => {
  it("renders every badge variant and supports slotted links", () => {
    const variants = ["default", "secondary", "destructive", "outline", "ghost", "link"] as const;
    render(<div>{variants.map((variant) => <Badge key={variant} variant={variant}>{variant}</Badge>)}<Badge asChild><a href="#target">slotted</a></Badge></div>);
    for (const variant of variants) expect(screen.getByText(variant).getAttribute("data-variant")).toBe(variant);
    expect(screen.getByRole("link", { name: "slotted" }).getAttribute("data-slot")).toBe("badge");
  });

  it("renders every button variant and size with native and slotted elements", () => {
    const variants = ["default", "outline", "secondary", "ghost", "destructive", "link"] as const;
    const sizes = ["default", "xs", "sm", "lg", "icon", "icon-xs", "icon-sm", "icon-lg"] as const;
    render(<div>{variants.map((variant) => <Button key={variant} variant={variant}>{variant}</Button>)}{sizes.map((size) => <Button key={size} size={size}>{size}</Button>)}<Button asChild><a href="#button">slotted</a></Button></div>);
    const buttons = Array.from(document.querySelectorAll('[data-slot="button"]'));
    for (const variant of variants) expect(buttons.find((button) => button.textContent === variant)?.getAttribute("data-variant")).toBe(variant);
    for (const size of sizes) expect(buttons.find((button) => button.getAttribute("data-size") === size)).toBeTruthy();
    expect(document.querySelector('a[data-slot="button"]')?.textContent).toBe("slotted");
  });

  it("renders every card part and card size", () => {
    render(<><Card><CardHeader><CardTitle>Title</CardTitle><CardDescription>Description</CardDescription><CardAction>Action</CardAction></CardHeader><CardContent>Content</CardContent><CardFooter>Footer</CardFooter></Card><Card size="sm">Small</Card></>);
    for (const text of ["Title", "Description", "Action", "Content", "Footer", "Small"]) expect(screen.getByText(text)).toBeTruthy();
    expect(document.querySelector('[data-slot="card"][data-size="sm"]')).toBeTruthy();
  });

  it("supports decorative and semantic separators plus skeleton props", () => {
    render(<><Separator /><Separator orientation="vertical" decorative={false} aria-label="section divider" /><Skeleton data-testid="loading" /></>);
    expect(screen.getByRole("separator", { name: "section divider" }).getAttribute("data-orientation")).toBe("vertical");
    expect(screen.getByTestId("loading").className).toContain("animate-pulse");
  });
});
