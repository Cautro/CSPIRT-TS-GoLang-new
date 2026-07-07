import { type ReactNode } from "react";

export interface TabsSwitcherItem<T extends string> {
    value: T;
    label: ReactNode;
    hidden?: boolean;
    disabled?: boolean;
    badge?: ReactNode;
}

interface TabsSwitcherProps<T extends string> {
    items: TabsSwitcherItem<T>[];
    value: T;
    onChange: (value: T) => void;
    className?: string;
}

export function TabsSwitcher<T extends string>({
                                                   items,
                                                   value,
                                                   onChange,
                                                   className = "",
                                               }: TabsSwitcherProps<T>) {
    const visibleItems = items.filter((item) => !item.hidden);

    return (
        <div className={`tabs-switcher ${className}`} role="tablist">
            {visibleItems.map((item) => {
                const isActive = item.value === value;

                return (
                    <button
                        key={item.value}
                        className={
                            isActive
                                ? "tabs-switcher__item tabs-switcher__item--active"
                                : "tabs-switcher__item"
                        }
                        type="button"
                        role="tab"
                        aria-selected={isActive}
                        disabled={item.disabled}
                        onClick={() => onChange(item.value)}
                    >
                        <span className="tabs-switcher__label">
                            {item.label}
                        </span>

                        {item.badge && (
                            <span className="tabs-switcher__badge">
                                {item.badge}
                            </span>
                        )}
                    </button>
                );
            })}
        </div>
    );
}