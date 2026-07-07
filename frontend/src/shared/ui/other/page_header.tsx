import { type ReactNode } from "react";
import {
    BurgerDrawerMenu,
    type BurgerDrawerMenuItem,
} from "../other/burger_menu.tsx";
import {useNavigate} from "react-router-dom";

interface PageHeaderProps {
    eyebrow?: ReactNode;
    title: ReactNode;
    description?: ReactNode;

    menuTitle?: string;
    menuItems?: BurgerDrawerMenuItem[];

    actions?: ReactNode;
    meta?: ReactNode;

    className?: string;
    hasBackButton?: boolean;
}

export function PageHeader({eyebrow, title, description, menuTitle = "Меню", menuItems = [], actions, meta, className = "", hasBackButton = false,}: PageHeaderProps) {
    const navigate = useNavigate();
    const hasMenu = menuItems.some((item) => !item.hidden);

    return (
        <header className={`page-header ${className}`}>
            <div className="page-header__content">
                {eyebrow && (
                    <p className="page-header__eyebrow">
                        {eyebrow}
                    </p>
                )}

                <h1 className="page-header__title">
                    {title}
                </h1>

                {description && (
                    <p className="page-header__description">
                        {description}
                    </p>
                )}

                {meta && (
                    <div className="page-header__meta">
                        {meta}
                    </div>
                )}
            </div>
            
            <div className="page-header__actions">
                {hasBackButton && (
                    <button className="btn btn--secondary" onClick={() => navigate(-1)}>
                        Назад
                    </button>
                )}
    
                {(actions || hasMenu) && (
                    <>
                        {actions}
    
    
                        {hasMenu && (
                            <BurgerDrawerMenu
                                title={menuTitle}
                                items={menuItems}
                                side="right"
                            />
                            
                        )}
                        
                    </>
                )}
            </div>
            
        </header>
    );
}