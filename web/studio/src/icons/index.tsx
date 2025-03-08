import { FC } from "react";
import { IconProps } from "./types";

export const IconPlayerPlayFilled: FC<IconProps> = ({
    width = 24,
    height = 24,
    stroke = 1.5,
    fill,
    style,
    className,
}) => {
    return (
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width={width}
            height={height}
            viewBox="0 0 24 24"
            fill={fill}
            strokeWidth={stroke}
            className={className}
            style={style}
        >
            <path stroke="none" d="M0 0h24v24H0z" fill="none" />
            <path d="M6 4v16a1 1 0 0 0 1.524 .852l13 -8a1 1 0 0 0 0 -1.704l-13 -8a1 1 0 0 0 -1.524 .852z" />
        </svg>
    );
};

export const IconPlayerStopFilled: FC<IconProps> = ({
    width = 24,
    height = 24,
    stroke = 1.5,
    fill,
    style,
    className,
}) => {
    return (
        <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            width={width}
            height={height}
            fill={fill}
            strokeWidth={stroke}
            className={className}
            style={style}
        >
            <path stroke="none" d="M0 0h24v24H0z" fill="none" />
            <path d="M17 4h-10a3 3 0 0 0 -3 3v10a3 3 0 0 0 3 3h10a3 3 0 0 0 3 -3v-10a3 3 0 0 0 -3 -3z" />
        </svg>
    );
};

export const IconPlaylistAd: FC<IconProps> = (props) => {
    return (
        <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            width={props.width}
            height={props.height}
            fill={props.fill}
            strokeWidth={props.stroke}
            className={props.className}
            style={props.style}
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
        >
            <path stroke="none" d="M0 0h24v24H0z" fill="none" />
            <path d="M19 8h-14" />
            <path d="M5 12h9" />
            <path d="M11 16h-6" />
            <path d="M15 16h6" />
            <path d="M18 13v6" />
        </svg>
    );
};
