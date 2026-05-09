"use client";

import { useState, forwardRef } from "react";
import Image from "next/image";
import { cn } from "@/lib/utils";
import { getAvatarSrc } from "@/lib/avatar";

interface AvatarProps extends React.HTMLAttributes<HTMLDivElement> {
  src?: string;
  alt?: string;
  fallbackInitial?: string;
  size?: number;
}

export const Avatar = forwardRef<HTMLDivElement, AvatarProps>(
  ({ src, alt, fallbackInitial, size = 36, className, onClick, ...rest }, ref) => {
    const [error, setError] = useState(false);
    const initial = fallbackInitial || alt?.[0]?.toUpperCase() || "?";
    const avatarSrc = getAvatarSrc(src, alt);
    const isDiceBear = avatarSrc.includes("dicebear");

    return (
      <div
        ref={ref}
        onClick={onClick}
        role={onClick ? "button" : undefined}
        tabIndex={onClick ? 0 : undefined}
        className={cn(
          "relative shrink-0 overflow-hidden rounded-lg shadow-[0_2px_6px_rgba(0,0,0,0.14)]",
          error
            ? "flex items-center justify-center bg-primary/15 text-sm font-semibold text-primary"
            : "",
          onClick && "cursor-pointer",
          className,
        )}
        style={{ width: size, height: size }}
        {...rest}
      >
        {error ? (
          <span>{initial}</span>
        ) : (
          <Image
            src={avatarSrc}
            alt={alt || ""}
            fill
            unoptimized={isDiceBear}
            onError={() => setError(true)}
            className="pointer-events-none object-cover"
          />
        )}
      </div>
    );
  }
);

Avatar.displayName = "Avatar";