import { useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";

export interface BurgerDrawerMenuItem {
    label: string;
    onClick: () => void | Promise<void>;
    disabled?: boolean;
    hidden?: boolean;
    danger?: boolean;
    primary?: boolean;
}

interface BurgerDrawerMenuProps {
    items: BurgerDrawerMenuItem[];
    title?: string;
    side?: "left" | "right";
}

export function BurgerDrawerMenu({
                                     items,
                                     title = "Меню",
                                     side = "right",
                                 }: BurgerDrawerMenuProps) {
    const [isOpen, setIsOpen] = useState(false);

    const visibleItems = useMemo(() => {
        return items.filter((item) => !item.hidden);
    }, [items]);

    useEffect(() => {
        if (!isOpen) {
            return;
        }

        function handleEscape(event: KeyboardEvent) {
            if (event.key === "Escape") {
                setIsOpen(false);
            }
        }

        document.addEventListener("keydown", handleEscape);
        document.body.style.overflow = "hidden";

        return () => {
            document.removeEventListener("keydown", handleEscape);
            document.body.style.overflow = "";
        };
    }, [isOpen]);

    const drawer = (
        <div
            className={isOpen ? "app-drawer app-drawer--open" : "app-drawer"}
            aria-hidden={!isOpen}
        >
            <div
                className="app-drawer__backdrop"
                onMouseDown={() => setIsOpen(false)}
            />

            <aside
                className={
                    side === "left"
                        ? "app-drawer__panel app-drawer__panel--left"
                        : "app-drawer__panel app-drawer__panel--right"
                }
                role="dialog"
                aria-modal="true"
                aria-labelledby="app-drawer-title"
            >
                <div className="app-drawer__header">
                    <h2 className="app-drawer__title" id="app-drawer-title">
                        {title}
                    </h2>

                    <button
                        className="app-drawer__close"
                        type="button"
                        onClick={() => setIsOpen(false)}
                        aria-label="Закрыть меню"
                    >
                        ×
                    </button>
                </div>

                <nav className="app-drawer__nav">
                    {visibleItems.map((item) => (
                        <button
                            key={item.label}
                            className={
                                item.danger
                                    ? "app-drawer__item app-drawer__item--danger"
                                    : item.primary
                                        ? "app-drawer__item app-drawer__item--primary"
                                        : "app-drawer__item"
                            }
                            type="button"
                            disabled={item.disabled}
                            onClick={async () => {
                                await item.onClick();
                                setIsOpen(false);
                            }}
                        >
                            {item.label}
                        </button>
                    ))}
                </nav>
            </aside>
        </div>
    );

    return (
        <>
            <button
                className="app-drawer-button"
                type="button"
                onClick={() => setIsOpen(true)}
                aria-label="Открыть меню"
                aria-expanded={isOpen}
            >
                <span className="app-drawer-button__line" />
                <span className="app-drawer-button__line" />
                <span className="app-drawer-button__line" />
            </button>

            {createPortal(drawer, document.body)}
        </>
    );
}